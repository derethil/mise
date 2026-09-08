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
      vendorHash = "sha256-xSVtxIohwVcDFkV5BgyqwAASsAwM4HNvwwYkTlSUd5A=";
      version = "0.2.0";
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

            packages = [pkgs.just pkgs.nodejs pkgs.fblog ollama];

            processes = {
              genkit = {
                exec = "genkit start -- mise genkit";
              };
              ollama = {
                exec = "${ollama}/bin/ollama serve";
                ready.http.get = {
                  path = "/api/version";
                  port = 11434;
                };
              };
            };

            scripts = {
              genkit.exec = ''
                npx --yes genkit-cli@latest "$@"
              '';

              mise.exec = ''
                cd "$DEVENV_ROOT" && go run . "$@"
              '';

              mlogs.exec = ''
                log_file="''${XDG_STATE_HOME:-$HOME/.local/state}/mise/mise.log"

                if [ "$1" = "--live" ] || [ "$1" = "-l" ]; then
                  tail -f "$log_file"
                else
                  cat "$log_file"
                fi | ${pkgs.fblog}/bin/fblog -d --main-line-format $'\n{{bold(fixed_size 19 fblog_timestamp)}} {{level_style (uppercase (fixed_size 5 fblog_level))}}:{{#if fblog_prefix}} {{bold(cyan fblog_prefix)}}{{/if}} {{fblog_message}}'
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
