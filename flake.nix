{
  description = "Groundwork — topology-aware GCP service scaffolder";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        # nix build
        packages.default = pkgs.buildGoModule {
          pname = "groundwork";
          version = "0.1.0";
          src = ./.;

          # vendor/ is committed — no hash needed.
          vendorHash = null;

          meta = with pkgs.lib; {
            description = "Topology-aware GCP service scaffolder";
            homepage = "https://github.com/philip-730/groundwork";
            license = licenses.mit;
            mainProgram = "groundwork";
          };
        };

        # nix develop
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            # Go toolchain
            go
            gopls
            gotools   # goimports, godoc, etc.

            # Required at runtime by registry sync
            git

            # Linting
            golangci-lint
          ];

          shellHook = ''
            echo "groundwork dev shell"
            echo "  go test ./...     — run tests"
            echo "  go build -o gw .  — build binary"
          '';
        };
      });
}
