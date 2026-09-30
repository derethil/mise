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

    goPkg = pkgs.go_1_27;
    runtimeDeps = [pkgs.ffmpeg pkgs.yt-dlp];

    # go-whisper's cgo LDFLAGS need a single libggml-cpu.so not dlopen'd microarch variants as packaged in nixpkgs
    whisperCppFor = pkgs: overrides:
      (pkgs.whisper-cpp.override overrides).overrideAttrs (old: {
        cmakeFlags =
          builtins.filter
          (f:
            !(pkgs.lib.any (needle: pkgs.lib.strings.hasInfix needle f) [
              "GGML_BACKEND_DL"
              "GGML_CPU_ALL_VARIANTS"
              "GGML_BACKEND_DIR"
            ]))
          old.cmakeFlags
          ++ ["-DGGML_BACKEND_DL=OFF" "-DGGML_CPU_ALL_VARIANTS=OFF"];
      });

    pkgWhisperCpp = whisperCppFor pkgs {};

    pkg = (pkgs.buildGoModule.override {go = goPkg;}) rec {
      buildInputs = [pkgWhisperCpp];
      env = {
        CGO_CFLAGS = "-I${pkgWhisperCpp}/include";
        CGO_LDFLAGS = "-L${pkgWhisperCpp}/lib";
      };
      ldflags = [
        "-s"
        "-w"
        "-X github.com/derethil/mise/cmd.version=${version}"
      ];
      meta.mainProgram = "mise";
      nativeBuildInputs = [pkgs.makeWrapper];
      pname = "mise";
      postFixup = ''
        wrapProgram $out/bin/mise --prefix PATH : ${pkgs.lib.makeBinPath runtimeDeps}
      '';
      src = ./.;
      vendorHash = "sha256-DaI6pYjlKPWGJrNa4swqBGz080MHryXPWzSQds3USvk=";
      version = "0.3.1";
    };

    mkShell = {
      ollama,
      whisperCpp,
    }:
      devenv.lib.mkShell {
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
              package = goPkg;
            };
            outputs = {
              mise = pkg;
            };
            packages = builtins.concatLists [
              [
                pkgs.just
                pkgs.nodejs
                pkgs.fblog
                ollama
                whisperCpp
              ]

              runtimeDeps
            ];
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
          })
        ];
      };
  in {
    # Pick the devshell build to use by setting MISE_DEVSHELL in .envrc.local
    devShells.${system} = rec {
      cpu = mkShell {
        inherit (unstabledPkgs) ollama;
        whisperCpp = whisperCppFor unstabledPkgs {};
      };
      cuda = mkShell {
        ollama = unstabledFreePkgs.ollama-cuda;
        whisperCpp = whisperCppFor unstabledFreePkgs {cudaSupport = true;};
      };
      default = cpu;
      rocm = mkShell {
        ollama = unstabledPkgs.ollama-rocm;
        whisperCpp = whisperCppFor unstabledPkgs {rocmSupport = true;};
      };
    };

    packages.${system} = {
      default = pkg;
      mise = pkg;
    };
  };
}
