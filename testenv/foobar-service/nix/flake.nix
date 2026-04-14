{
  description = "foobar-service";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in {
        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go
            gopls
            gotools
            google-cloud-sdk
            terraform
          ];

          shellHook = ''
            echo "dev shell: foobar-service"
            echo ""
            echo "  gcloud auth application-default login"
            echo "  terraform -chdir=terraform init"
          '';
        };
      });
}
