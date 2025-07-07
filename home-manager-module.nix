{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.steam-shortcuts;

  # TODO: Support: pkgs.stdenv.hostPlatform.isDarwin
  # This hardcoded steam path seems pretty consistent
  steamConfDir = "${config.xdg.configHome}/../.steam/steam";
  userConfigDir = "${steamConfDir}/userdata/${cfg.steamUserId}/config";
in
  with lib; {
    options.services.steam-shortcuts = {
      enable = mkEnableOption "Steam shortcuts management";

      package = mkOption {
        type = types.package;
        default = pkgs.callPackage ./default.nix {};
        defaultText = literalExpression "inputs.json2steamshortcut.packages.\${stdenv.hostPlatform.system}.json2steamshortcut";
      };
      steamUserId = mkOption {
        type = types.str;
        description = ''
          The Steam user ID is the numeric identifier for your Steam account.
          It can be found in the Steam userdata directory, typically located at
          ~/.local/share/Steam/userdata/<steamUserId>/
          or
          ~/.steam/steam/userdata/<steamUserId>/

          You can also find your Steam user ID by checking the URL of your profile page
          in your browser or in your Steam client.
        '';

        example = "158842264";
      };

      # NOTE: This configuration is formatted for use with json2steamshortcut
      #       and will be converted to JSON and piped via stdout to json2steamshortcut.
      #
      #       Notably, Steam uses lowercase keys for most of these options,
      #       excepting StartDir and LaunchOptions
      shortcuts = mkOption {
        type = types.listOf (types.submodule {
          options = {
            AppName = mkOption {
              type = types.str;
              description = "Name of the application as it appears in Steam";
            };
            Exe = mkOption {
              type = types.str;
              description = "Path to the executable";
            };
            StartDir = mkOption {
              type = types.str;
              default = "";
              description = "Starting directory for the application";
            };
            LaunchOptions = mkOption {
              type = types.str;
              default = "";
              description = "Launch options for the application";
            };
            Icon = mkOption {
              type = types.str;
              default = "";
              description = "Path to icon file";
            };
            Tags = mkOption {
              type = types.listOf types.str;
              default = [];
              description = "Tags for the shortcut";
            };
          };
        });
        default = [];
        description = "List of shortcuts to create";
        example = literalExpression ''
          [
            {
              AppName = "Performous";
              Exe = "''${pkgs.performous}/bin/performous";
            }
            {
              AppName = "Firefox";
              Exe = "''${pkgs.firefox}/bin/firefox";
              LaunchOptions = "--new-window";
            }
          ]
        '';
      };
    };

    config = mkIf cfg.enable {
      assertions = [
        {
          # TODO: Use simple type checking if possible /[0-9]+/
          assertion = cfg.steamUserId != "";
          message = "services.steam-shortcuts requires steamUserId to be set";
        }
        {
          assertion = cfg.shortcuts != [];
          message = "services.steam-shortcuts expects at least one shortcut to be defined";
        }
      ];

      # Attempt to let shortcuts.vdf be writable by steam
      # even if for now, we don't persist the changes.

      # Taken from https://github.com/nix-community/home-manager/blob/ea24675e4f4f4c494ccb04f6645db2a394d348ee/modules/programs/vscode/default.nix#L354C5-L386C7
      # Description: https://github.com/nix-community/home-manager/blob/ea24675e4f4f4c494ccb04f6645db2a394d348ee/modules/home-environment.nix#L429
      # Not sure if this is the recommended approach for a relatively easy file modification,
      # worth to mention as well, this is destructive as it overwrites any existing shortcuts.vdf

      # Create or overwrite the shortcuts.vdf file
      home.activation.steam-shortcuts = let
        json = builtins.toJSON cfg.shortcuts;
        vdf = pkgs.runCommandLocal "shortcuts.vdf" {
          nativeBuildInputs = [cfg.package];
        } "echo '${json}' | json2steamshortcut > $out";
      in
        lib.hm.dag.entryAfter ["writeBoundary"] ''
          # TODO: Is it a good practice to create folder if it doesn't exist?
          #       Should only be applicable for systems that haven't yet started
          #       Steam for the first time as the user designated by `cfg.steamUserId`.
          #
          # - Skips creating shortcuts if '.steam/steam/userdata' doesn't exist
          [ -d "${userConfigDir}" ] || {
            [ -d "${steamConfDir}/userdata" ] || {
              verboseEcho "${steamConfDir}/userdata doesn't exist, skipped creating ${userConfigDir}/shortcuts.vdf" ;
              exit 0 ;
            }
            verboseEcho "Creating ${userConfigDir}" ;
            run mkdir -p "${userConfigDir}" ;
          }
          verboseEcho "Writing ${userConfigDir}/shortcuts.vdf"
          run cat "${vdf}" > "${userConfigDir}/shortcuts.vdf"
        '';
    };
  }
