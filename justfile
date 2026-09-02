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
