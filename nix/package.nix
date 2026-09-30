{
  perSystem = {
    mise,
    pkgs,
    ...
  }: let
    whisperCpp = mise.whisperCppFor pkgs {};

    package = (pkgs.buildGoModule.override {go = mise.goPackage;}) rec {
      buildInputs = [whisperCpp];
      env = {
        CGO_CFLAGS = "-I${whisperCpp}/include";
        CGO_LDFLAGS = "-L${whisperCpp}/lib";
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
        wrapProgram $out/bin/mise --prefix PATH : ${pkgs.lib.makeBinPath mise.runtimeDependencies}
      '';
      src = ../.;
      vendorHash = "sha256-DaI6pYjlKPWGJrNa4swqBGz080MHryXPWzSQds3USvk=";
      version = "0.3.1";
    };
  in {
    packages = {
      default = package;
      mise = package;
    };
  };
}
