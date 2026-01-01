{
  description = "golazo";

  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = inputs @ {flake-parts, ...}:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      perSystem = {pkgs, ...}: {
        packages.default = pkgs.callPackage ./. {inherit pkgs;};

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go_1_24
            golangci-lint-langserver
            golangci-lint
            gopls
            go-tools
            gotools
          ];

          shellHook =
            # sh
            ''
              export CGO_ENABLED=0
            '';
        };
      };
    };
}
