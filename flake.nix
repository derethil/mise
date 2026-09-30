{
  description = "mise";

  inputs = {
    devenv.url = "github:cachix/devenv";
    flake-parts.follows = "devenv/flake-parts";
    nixpkgs.url = "github:cachix/devenv-nixpkgs/rolling";
    unstable.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = ["x86_64-linux"];

      imports = [
        ./nix/shared.nix
        ./nix/package.nix
        ./nix/dev-shells.nix
      ];
    };
}
