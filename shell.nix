{ sources ? import ./nix/sources.nix
, pkgs ? import sources.nixpkgs {}
, frameworks ? pkgs.darwin.apple_sdk.frameworks
}:

pkgs.mkShell {
  nativeBuildInputs = [
    pkgs.go_1_22
    pkgs.nodejs

    pkgs.pkg-config

    # for node-canvas
    pkgs.pixman
    pkgs.python39
    pkgs.cairo
    pkgs.pango
    # end node-canvas

    frameworks.Security
    frameworks.Cocoa
    frameworks.WebKit
    frameworks.UniformTypeIdentifiers
    frameworks.ForceFeedback
    frameworks.Kernel
  ];
}
