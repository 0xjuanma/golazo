{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "golazo";
  version = "0.32.0";
  vendorHash = "sha256-m0h/pgcom/P6NdHvuhhiAhw1kQqV+JHM7Tb8AJ37eD8=";

  subPackages = ["."];

  src = builtins.path {
    path = ./.;
    name = "source";
  };
}
