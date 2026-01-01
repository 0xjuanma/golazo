{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "golazo";
  version = "0.8.0";
  vendorHash = "sha256-qtDFhudYgBwvxfzDiSWB4dG6m0e7kbsmt5BU2fEoOGw=";

  subPackages = ["."];

  src = builtins.path {
    path = ./.;
    name = "source";
  };
}
