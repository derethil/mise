{inputs, ...}: {
  perSystem = {
    pkgs,
    system,
    ...
  }: let
    unstablePackages = inputs.unstable.legacyPackages.${system};
    unstableUnfreePackages = import inputs.unstable {
      inherit system;
      config.allowUnfree = true;
    };

    goPackage = pkgs.go_1_27;
    runtimeDependencies = [pkgs.ffmpeg pkgs.yt-dlp];

    # go-whisper's cgo LDFLAGS need a single libggml-cpu.so, not
    # dlopen'd microarchitecture variants as packaged in nixpkgs.
    # nixpkgs also passes the old GGML_HIPBLAS instead of GGML_HIP.
    whisperCppFor = packageSet: overrides:
      (packageSet.whisper-cpp.override overrides).overrideAttrs (old: {
        cmakeFlags =
          builtins.filter
          (flag:
            !(packageSet.lib.any (needle: packageSet.lib.strings.hasInfix needle flag) [
              "GGML_BACKEND_DL"
              "GGML_CPU_ALL_VARIANTS"
              "GGML_BACKEND_DIR"
              "GGML_HIPBLAS"
            ]))
          old.cmakeFlags
          ++ ["-DGGML_BACKEND_DL=OFF" "-DGGML_CPU_ALL_VARIANTS=OFF"]
          ++ packageSet.lib.optionals (overrides.rocmSupport or false) ["-DGGML_HIP=ON"];
      });
  in {
    _module.args.mise = {
      inherit
        goPackage
        runtimeDependencies
        unstablePackages
        unstableUnfreePackages
        whisperCppFor
        ;
    };
  };
}
