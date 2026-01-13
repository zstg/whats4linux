#!/usr/bin/env nix-shell
#!nix-shell -p coreutils curl common-updater-scripts jq prefetch-npm-deps go
# shellcheck shell=bash

set -euo pipefail

# version=$(curl -s "https://api.github.com/repos/lugvitc/whats4linux/releases/latest" | jq -r ".tag_name" | tr -d v)

# if [[ "$UPDATE_NIX_OLD_VERSION" == "$version" ]]; then
  # echo "Already up to date!"
  # exit 0
# fi

pushd "$(mktemp -d)"
curl -s "https://raw.githubusercontent.com/lugvitc/whats4linux/master/frontend/package-lock.json" -o frontend-package-lock.json
newFrontendDepsHash=$("${pkgs.prefetch-npm-deps}/bin/prefetch-npm-deps" frontend-package-lock.json)
popd

PACKAGE_ROOT=$(dirname "${BASH_SOURCE[0]}")

# Get current vendorHash from package.nix
oldVendorHash=$(grep -o 'vendorHash = "[^"]*"' "$PACKAGE_ROOT"/package.nix | sed 's/vendorHash = "//;s/"//g')

# Update frontend deps hash
sed -i "s|npmDepsHash = \".*\"|npmDepsHash = \"$newFrontendDepsHash\"|g" "$PACKAGE_ROOT"/package.nix

echo "Updated npmDepsHash to: $newFrontendDepsHash"

if [ -n "${1-}" ] && [ "$1" = "--update-vendor" ]; then
  echo "Updating vendorHash..."
  pushd "$PACKAGE_ROOT"
  # go mod tidy && go mod vendor # Will this work?
  # Calculate new vendor hash by attempting to build with null hash
  tempHash=$(nix-build --expr '
    with import <nixpkgs> {};
    let 
      pkg = callPackage ./package.nix { vendorHash = null; };
    in
    (builtins.unsafeDiscardStringContext pkg.src.vendorDir.hash)
  2>/dev/null || echo "Failed to get vendor hash")
  
  if [ -n "$tempHash" ] && [ "$tempHash" != "Failed to get vendor hash" ]; then
    newVendorHash="sha256-$tempHash"
    sed -i "s|vendorHash = \".*\"|vendorHash = \"$newVendorHash\"|g" "$PACKAGE_ROOT"/package.nix
    echo "Updated vendorHash to: $newVendorHash"
  else
    echo "Failed to calculate new vendorHash. Please run: nix-build --expr 'with import <nixpkgs> {}; callPackage ./package.nix {}' and use the provided hash."
  fi
  popd
else
  echo "To update vendorHash, run: $0 --update-vendor"
fi
