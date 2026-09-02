{
  description = "mise";

  inputs = {
    devenv.url = "github:cachix/devenv";
    nixpkgs.url = "github:cachix/devenv-nixpkgs/rolling";
  };

  outputs = {
    devenv,
    nixpkgs,
    ...
  } @ inputs: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};

    pkg = pkgs.buildGoModule {
      meta.mainProgram = "mise";
      pname = "mise";
      src = ./.;
      vendorHash = "sha256-s6wK+EG43Zb99/gylTmxz3seKpBscONDp7nYZqhQxIE=";
      version = "0.1.0";
    };
  in {
    devShells.${system}.default = devenv.lib.mkShell {
      inherit inputs pkgs;

      modules = [
        {
          languages.go.enable = true;
          outputs = {
            mise = pkg;
          };
          packages = [pkgs.just];
          scripts.mise.exec = ''
            cd "$DEVENV_ROOT" && go run . "$@"
          '';
        }
      ];
    };

    packages.${system} = {
      default = pkg;
      mise = pkg;
    };
  };
}
