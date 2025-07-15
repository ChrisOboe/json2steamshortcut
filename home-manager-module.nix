{
  config,
  lib,
  pkgs,
  ...
}: let
  cfg = config.services.steam-shortcuts;

  # This hardcoded steam path seems pretty consistent, at least for Linux
  steamConfDir =
    if pkgs.stdenv.hostPlatform.isDarwin
    then "${config.home.homeDirectory}/Library/Application Support/Steam"
    else "${config.home.homeDirectory}/.steam/steam";
  userConfigDir = "${steamConfDir}/userdata/${builtins.toString cfg.steamUserId}/config";
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
        type = types.int;
        description = ''
          The Steam user ID is the numeric identifier for your Steam account.
          It can be found in the Steam userdata directory, typically located at
          ~/.local/share/Steam/userdata/<steamUserId>/
          or
          ~/.steam/steam/userdata/<steamUserId>/

          You can also find your Steam user ID by checking the URL of your profile page
          in your browser or in your Steam client.
        '';

        example = 158842264;
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
          assertion = cfg.shortcuts != [];
          message = "services.steam-shortcuts expects at least one shortcut to be defined";
        }
        {
          assertion = config.home.homeDirectory != "";
          message = "Home directory could not be determined";
        }
      ];

      # Create shortcuts.vdf file
      home.file."${userConfigDir}/shortcuts.vdf" = let
        json = builtins.toJSON cfg.shortcuts;
        vdf = pkgs.runCommandLocal "shortcuts.vdf" {
          nativeBuildInputs = [cfg.package];
        } "echo '${json}' | json2steamshortcut > $out";
      in {
        source = vdf;
        force = true;
      };
    };
  }
