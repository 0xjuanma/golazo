{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "golazo";
  version = "0.28.0";
  vendorHash = "sha256-8p3JyLcFcHRAYoQn6/u43T4YsyVWzXIAYjLbzP8O584=";

  subPackages = ["."];

  src = builtins.path {
    path = ./.;
    name = "source";
  };
}
