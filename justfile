[private]
default:
    @just --list

check:
    #!/usr/bin/env bash

    echo "Running tests..."
    go vet ./...
    go test ./...

    echo "Verifying build..."
    go build -o mise
    rm -f mise

# Run unit tests, or the Tandoor integration suite against a specific version
[arg("selector", pattern="^(integration)?$", help="Pass 'integration' to run the Tandoor integration suite instead of the unit tests")]
[arg("version", pattern="^([0-9]+[.][0-9]+[.][0-9]+)?$", help="Tandoor version the integration suite runs against (e.g. 2.6.13)")]
test selector="" version="":
    #!/usr/bin/env bash
    set -euo pipefail

    if [ -z "{{ selector }}" ]; then
        if [ -n "{{ version }}" ]; then
            echo "Error: a version requires the 'integration' selector: just test integration {{ version }}" >&2
            exit 1
        fi

        go test ./...
        exit 0
    fi

    if [ -z "{{ version }}" ]; then
        echo "Error: 'just test integration' requires a Tandoor version, e.g. just test integration 2.6.13" >&2
        exit 1
    fi

    TANDOOR_VERSION="{{ version }}" go test -tags=integration -v ./integration/tandoor

# Validate and promote a newer Tandoor maximum after its integration suite passes
[arg("version", pattern="^[0-9]+[.][0-9]+[.][0-9]+$", help="Tandoor version to validate and support (e.g. 2.6.15)")]
support version:
    #!/usr/bin/env bash
    set -euo pipefail

    just test integration "{{ version }}"
    just update-support-bound maximum "{{ version }}"

# Deprecate versions below an already-supported Tandoor version
[arg("version", pattern="^[0-9]+[.][0-9]+[.][0-9]+$", help="Lowest Tandoor version to continue supporting (e.g. 2.6.15)")]
deprecate version:
    #!/usr/bin/env bash
    set -euo pipefail

    just update-support-bound minimum "{{ version }}"

[private]
update-support-bound bound version:
    #!/usr/bin/env bash
    set -euo pipefail

    version_file="internal/tandoor/version.go"
    case "{{ bound }}" in
        maximum)
            constant="MaxSupportedVersion"
            commit_message="chore(tandoor): bump supported version to {{ version }}"
            ;;
        minimum)
            constant="MinSupportedVersion"
            commit_message="chore(tandoor): deprecate versions before {{ version }}"
            ;;
        *)
            echo "Error: unknown support bound {{ bound }}" >&2
            exit 1
            ;;
    esac

    minimum="$(sed -nE 's/^[[:space:]]*MinSupportedVersion = "([0-9]+\.[0-9]+\.[0-9]+)"$/\1/p' "$version_file")"
    maximum="$(sed -nE 's/^[[:space:]]*MaxSupportedVersion = "([0-9]+\.[0-9]+\.[0-9]+)"$/\1/p' "$version_file")"
    if [ -z "$minimum" ] || [ -z "$maximum" ]; then
        echo "Error: could not find the supported version range in $version_file" >&2
        exit 1
    fi

    IFS=. read -r candidate_major candidate_minor candidate_patch <<< "{{ version }}"
    IFS=. read -r minimum_major minimum_minor minimum_patch <<< "$minimum"
    IFS=. read -r maximum_major maximum_minor maximum_patch <<< "$maximum"

    if [ "{{ bound }}" = maximum ] && (( candidate_major < maximum_major ||
        (candidate_major == maximum_major && candidate_minor < maximum_minor) ||
        (candidate_major == maximum_major && candidate_minor == maximum_minor && candidate_patch <= maximum_patch) )); then
        echo "Tandoor {{ version }} is not newer than the supported maximum $maximum; support range unchanged."
        exit 0
    fi

    if [ "{{ bound }}" = minimum ] && (( candidate_major > maximum_major ||
        (candidate_major == maximum_major && candidate_minor > maximum_minor) ||
        (candidate_major == maximum_major && candidate_minor == maximum_minor && candidate_patch > maximum_patch) )); then
        echo "Error: Tandoor {{ version }} is newer than the supported maximum $maximum" >&2
        exit 1
    fi

    if [ "{{ bound }}" = minimum ] && (( candidate_major < minimum_major ||
        (candidate_major == minimum_major && candidate_minor < minimum_minor) ||
        (candidate_major == minimum_major && candidate_minor == minimum_minor && candidate_patch <= minimum_patch) )); then
        echo "Tandoor {{ version }} does not raise the supported minimum $minimum; support range unchanged."
        exit 0
    fi

    matches="$(grep -Ec "^[[:space:]]*$constant = \"" "$version_file")"
    if [ "$matches" -ne 1 ]; then
        echo "Error: expected exactly one $constant in $version_file, found $matches" >&2
        exit 1
    fi

    current="$minimum"
    if [ "{{ bound }}" = maximum ]; then
        current="$maximum"
    fi

    sed -i "s|$constant = \"$current\"|$constant = \"{{ version }}\"|" "$version_file"
    go test ./internal/tandoor -run 'TestVersionSupported|TestCompareVersions'

    git add "$version_file"
    git commit --only "$version_file" -m "$commit_message"

# Run tests with coverage, printing per-func % coverage
[arg("html", long="html", short="h", value="true", help="Open an HTML coverage report instead of printing per-func coverage")]
coverage html="false":
    #!/usr/bin/env bash
    set -euo pipefail

    go test ./... -coverprofile=coverage.out

    if [ "{{ html }}" = "true" ]; then
        go tool cover -html=coverage.out
    else
        go tool cover -func=coverage.out
    fi

    rm -f coverage.out

# Recompute flake.nix's vendorHash after a go.mod/go.sum change
update-vendor-hash:
    #!/usr/bin/env bash
    set -euo pipefail

    sed -i 's|vendorHash = ".*";|vendorHash = pkgs.lib.fakeHash;|' flake.nix

    output="$(nix build .#mise 2>&1)" || true
    hash="$(echo "$output" | grep 'got:' | awk '{print $NF}')"

    if [ -z "$hash" ]; then
        echo "No hash mismatch found in nix build output - vendorHash may already be correct, or the build failed for another reason:"
        echo "$output"
        exit 1
    fi

    sed -i "s|vendorHash = pkgs.lib.fakeHash;|vendorHash = \"$hash\";|" flake.nix

    nix build .#mise
    echo "vendorHash updated to $hash"

# Validate, tag, and push a release for the given version
[arg("version", pattern="^[0-9]+[.][0-9]+[.][0-9]+$", help="Version to release, without the leading v (e.g. 0.2.0)")]
release version:
    #!/usr/bin/env bash
    set -euo pipefail

    tag="v{{ version }}"

    if ! command -v gh &> /dev/null; then
        echo "Error: 'gh' (GitHub CLI) is not installed."
        exit 1
    fi

    if ! git diff-index --quiet HEAD --; then
        echo "Error: You have uncommitted changes. Please commit or stash them first."
        exit 1
    fi

    current_branch="$(git branch --show-current)"
    if [ "$current_branch" != "main" ]; then
        echo "Error: You are on branch '$current_branch'. Releases must be performed from 'main'."
        exit 1
    fi

    if git rev-parse "$tag" >/dev/null 2>&1; then
        echo "Error: Tag $tag already exists."
        exit 1
    fi

    echo "Running tests..."
    go vet ./...
    go test ./...

    echo "Verifying the flake still builds..."
    nix build .#mise
    rm -f result

    echo "Bumping version to {{ version }} in flake.nix..."
    sed -i 's/version = "[0-9][0-9.]*";/version = "{{ version }}";/' flake.nix
    if git diff --quiet -- flake.nix; then
        echo "flake.nix is already at version {{ version }}, skipping commit."
    else
        git add flake.nix
        git commit -m "chore: bump version to {{ version }}"
    fi

    echo "Tagging $tag..."
    git tag -a "$tag" -m "Release $tag"

    echo "Pushing main and $tag..."
    git push origin main
    git push origin "$tag"

    echo "Waiting for the release workflow..."
    run_id=""
    for _ in $(seq 30); do
        run_id="$(gh run list --workflow=release.yml --branch "$tag" --limit 1 --json databaseId --jq '.[0].databaseId')"
        [ -n "$run_id" ] && break
        sleep 5
    done
    if [ -z "$run_id" ]; then
        echo "Error: no release workflow run found for $tag."
        exit 1
    fi
    gh run watch "$run_id" --exit-status

    echo "Released $tag"
