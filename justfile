[private]
default:
    @just --list

# Recompute flake.nix's vendorHash after go.mod/go.sum change.
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

# Validate, tag, and push a release for the given version (e.g. `just release 0.2.0`).
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
    git add flake.nix
    git commit -m "chore: bump version to {{ version }}"

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
