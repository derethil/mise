{
  description = "mise";

  inputs = {
    devenv.url = "github:cachix/devenv";
    nixpkgs.url = "github:cachix/devenv-nixpkgs/rolling";
    unstable.url = "github:nixos/nixpkgs/nixos-unstable";
  };

  outputs = {
    devenv,
    nixpkgs,
    unstable,
    ...
  } @ inputs: let
    system = "x86_64-linux";
    pkgs = nixpkgs.legacyPackages.${system};

    unstabledPkgs = unstable.legacyPackages.${system};
    unstabledFreePkgs = import unstable {
      inherit system;
      config.allowUnfree = true;
    };

    pkg = pkgs.buildGoModule rec {
      ldflags = [
        "-s"
        "-w"
        "-X github.com/derethil/mise/cmd.version=${version}"
      ];
      meta.mainProgram = "mise";
      pname = "mise";
      src = ./.;
      vendorHash = "sha256-s6wK+EG43Zb99/gylTmxz3seKpBscONDp7nYZqhQxIE=";
      version = "0.1.0";
    };

    mkShell = ollama:
      devenv.lib.mkShell {
        inherit inputs pkgs;

        modules = [
          {
            languages.go.enable = true;

            outputs = {
              mise = pkg;
            };

            packages = [pkgs.just pkgs.nodejs ollama];

            processes.ollama = {
              exec = "${ollama}/bin/ollama serve";
              ready.http.get = {
                path = "/api/version";
                port = 11434;
              };
            };

            scripts = {
              genkit.exec = ''
                npx --yes genkit-cli@latest "$@"
              '';

              mise.exec = ''
                cd "$DEVENV_ROOT" && go run . "$@"
              '';
            };
          }
        ];
      };
  in {
    # Pick the ollama build to use by setting MISE_DEVSHELL in .envrc.local
    devShells.${system} = {
      cpu = mkShell unstabledPkgs.ollama;
      cuda = mkShell unstabledFreePkgs.ollama-cuda;
      default = mkShell unstabledPkgs.ollama;
      rocm = mkShell unstabledPkgs.ollama-rocm;
    };

    packages.${system} = {
      default = pkg;
      mise = pkg;
    };
  };
}
