{inputs, ...}: {
  perSystem = {
    config,
    mise,
    pkgs,
    ...
  }: let
    misePackage = config.packages.mise;

    mkShell = {
      ollama,
      whisperCpp,
    }:
      inputs.devenv.lib.mkShell {
        inherit inputs pkgs;

        modules = [
          ({config, ...}: {
            env = {
              CGO_CFLAGS = "-I${whisperCpp}/include";
              CGO_LDFLAGS = "-L${whisperCpp}/lib";
            };

            git-hooks.hooks = {
              gofmt.enable = true;

              gotest = {
                enable = true;
                excludes = ["^integration/"];
              };

              govet = {
                enable = true;
                excludes = ["^integration/"];
              };

              govet-integration = {
                enable = true;
                entry = "${config.languages.go.package}/bin/go vet -tags=integration ./integration/...";
                files = "^integration/.*\\.go$";
                name = "govet (integration)";
                pass_filenames = false;
              };
            };

            git-hooks.tools.go = config.languages.go.package;

            languages.go = {
              enable = true;
              package = mise.goPackage;
            };

            outputs.mise = misePackage;

            packages = builtins.concatLists [
              [
                pkgs.just
                pkgs.nodejs
                pkgs.fblog
                ollama
                whisperCpp
              ]

              mise.runtimeDependencies
            ];

            processes = {
              genkit.exec = "genkit start -- mise genkit";
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
          })
        ];
      };
  in {
    # Pick the dev shell build by setting MISE_DEVSHELL in .envrc.local.
    devShells = rec {
      cpu = mkShell {
        inherit (mise.unstablePackages) ollama;
        whisperCpp = mise.whisperCppFor mise.unstablePackages {};
      };

      cuda = mkShell {
        ollama = mise.unstableUnfreePackages.ollama-cuda;
        whisperCpp = mise.whisperCppFor mise.unstableUnfreePackages {cudaSupport = true;};
      };

      default = cpu;

      rocm = mkShell {
        ollama = mise.unstablePackages.ollama-rocm;
        whisperCpp = mise.whisperCppFor mise.unstablePackages {rocmSupport = true;};
      };
    };
  };
}
