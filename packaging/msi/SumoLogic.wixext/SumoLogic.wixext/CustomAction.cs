using System;
using System.IO;
using System.Text.RegularExpressions;
using System.Windows.Forms;
using WixToolset.Dtf.WindowsInstaller;

namespace SumoLogic.wixext
{
    public class CustomActions
    {
        // WiX error codes
        private const short ecMissingCustomActionData = 9001;
        private const short ecConfigError = 9002;

        // WiX property names
        private const string pCommonConfigPath = "CommonConfigPath";
        private const string pSumoLogicConfigPath = "SumoLogicConfigPath";
        private const string pInstallationToken = "InstallationToken";
        private const string pInstallToken = "InstallToken";
        private const string pTags = "Tags";
        private const string pApi = "Api";
        private const string pRemotelyManaged = "RemotelyManaged";
        private const string pEphemeral = "Ephemeral";
        private const string pConfigFolder = "ConfigFolder";
        private const string pOpAmpFolder = "OpAmpFolder";
        private const string pConfigFragmentsFolder = "ConfigFragmentsFolder";

        // WiX features
        private const string fRemotelyManaged = "REMOTELYMANAGED";

        [CustomAction]
        public static ActionResult UpdateConfig(Session session)
        {
            Logger logger = new Logger("UpdateConfig", session);
            logger.Log("Begin");

            // Validate presence of required WiX properties
            if (!session.CustomActionData.ContainsKey(pCommonConfigPath))
            {
                ShowErrorMessage(session, ecMissingCustomActionData, pCommonConfigPath);
                return ActionResult.Failure;
            }
            if (!session.CustomActionData.ContainsKey(pInstallToken) && !session.CustomActionData.ContainsKey(pInstallationToken))
            {
                ShowErrorMessage(session, ecMissingCustomActionData, pInstallationToken);
                return ActionResult.Failure;
            }
            if (!session.CustomActionData.ContainsKey(pTags))
            {
                ShowErrorMessage(session, ecMissingCustomActionData, pTags);
                return ActionResult.Failure;
            }

            var commonConfigPath = session.CustomActionData[pCommonConfigPath];
            var sumoLogicConfigPath = session.CustomActionData[pSumoLogicConfigPath];
            var tags = session.CustomActionData[pTags];

            var installationToken = "";
            if (session.CustomActionData.ContainsKey(pInstallationToken) && session.CustomActionData[pInstallationToken] != "")
            {
                installationToken = session.CustomActionData[pInstallationToken];
            }
            else if (session.CustomActionData.ContainsKey(pInstallToken))
            {
                installationToken = session.CustomActionData[pInstallToken];
            }

            var remotelyManaged = (session.CustomActionData.ContainsKey(pRemotelyManaged) && session.CustomActionData[pRemotelyManaged] == "true");
            var ephemeral = (session.CustomActionData.ContainsKey(pEphemeral) && session.CustomActionData[pEphemeral] == "true");
            var opAmpFolder = session.CustomActionData[pOpAmpFolder];
            var api = session.CustomActionData[pApi];

            if (remotelyManaged && string.IsNullOrEmpty(opAmpFolder))
            {
                ShowErrorMessage(session, ecMissingCustomActionData, pOpAmpFolder);
                return ActionResult.Failure;
            }

            // Load config from disk and replace values
            Config config = new Config { InstallationToken = installationToken, RemotelyManaged = remotelyManaged, Ephemeral = ephemeral,
                OpAmpFolder = opAmpFolder, Api = api };
            config.SetCollectorFieldsFromTags(tags);

            var configFile = remotelyManaged ? sumoLogicConfigPath : commonConfigPath;
            try
            {
                ConfigUpdater configUpdater = new ConfigUpdater(new StreamReader(configFile));
                configUpdater.Update(config);
                configUpdater.Save(new StreamWriter(configFile));
            }
            catch (Exception e)
            {
                ShowErrorMessage(session, ecConfigError, e.Message);
                return ActionResult.Failure;
            }

            logger.Log("End");

            return ActionResult.Success;
        }

        [CustomAction]
        public static ActionResult PopulateServiceArguments(Session session)
        {
            try
            {
                var configFolder = session.GetTargetPath(pConfigFolder);
                var configFragmentsFolder = session.GetTargetPath(pConfigFragmentsFolder);
                var remotelyManaged = session.Features.Contains(fRemotelyManaged) && session.Features[fRemotelyManaged].RequestState == InstallState.Local;

                if (remotelyManaged)
                {
                    session["SERVICEARGUMENTS"] = " --remote-config \"opamp:" + configFolder + "sumologic.yaml\"";
                }
                else
                {
                    session["SERVICEARGUMENTS"] = " --config \"" + configFolder + "sumologic.yaml\" --config \"glob:" + configFragmentsFolder + "*.yaml\"";
                }
            }
            catch (Exception e)
            {
                ShowErrorMessage(session, ecConfigError, e.Message + e.StackTrace);
                return ActionResult.Failure;
            }

            return ActionResult.Success;
        }

        private static void ShowErrorMessage(Session session, short errCode, string key)
        {
            Record record = new Record(3);
            record.SetInteger(1, errCode);
            record.SetString(2, key);
            session.Message(InstallMessage.Error | (InstallMessage)MessageIcon.Error, record);
            record.Close();
        }
    }
}
