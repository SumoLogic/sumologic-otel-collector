using System;
using System.IO;
using System.Text.RegularExpressions;
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
        private const string pInstallationToken = "InstallationToken";
        private const string pInstallToken = "InstallToken";
        private const string pTags = "Tags";

        [CustomAction]
        public static ActionResult UpdateConfig(Session session)
        {
            Logger logger = new Logger("UpdateConfig", session);
            logger.Log("Begin");

            // Validate presence of required WiX properties
            if (!session.CustomActionData.ContainsKey(pCommonConfigPath))
            {
                showErrorMessage(session, ecMissingCustomActionData, pCommonConfigPath);
                return ActionResult.Failure;
            }
            if (!session.CustomActionData.ContainsKey(pInstallToken) && !session.CustomActionData.ContainsKey(pInstallationToken))
            {
                showErrorMessage(session, ecMissingCustomActionData, pInstallationToken);
                return ActionResult.Failure;
            }
            if (!session.CustomActionData.ContainsKey(pTags))
            {
                showErrorMessage(session, ecMissingCustomActionData, pTags);
                return ActionResult.Failure;
            }

            var commonConfigPath = session.CustomActionData[pCommonConfigPath];
            var tags = session.CustomActionData[pTags];

            var installationToken = "";
            if (session.CustomActionData.ContainsKey(pInstallationToken) && session.CustomActionData[pInstallationToken] != "") {
                installationToken = session.CustomActionData[pInstallationToken];
            } else if (session.CustomActionData.ContainsKey(pInstallToken)){
                installationToken = session.CustomActionData[pInstallToken];
            }

            // Load config from disk and replace values
            Config config = new Config();
            config.InstallationToken = installationToken;
            config.SetCollectorFieldsFromTags(tags);

            try
            {
                ConfigUpdater configUpdater = new ConfigUpdater(new StreamReader(commonConfigPath));
                configUpdater.Update(config);
                configUpdater.Save(new StreamWriter(commonConfigPath));
            }
            catch (Exception e)
            {
                showErrorMessage(session, ecConfigError, e.Message);
                return ActionResult.Failure;
            }

            logger.Log("End");

            return ActionResult.Success;
        }

        private static void showErrorMessage(Session session, short errCode, string key)
        {
            Record record = new Record(3);
            record.SetInteger(1, errCode);
            record.SetString(2, key);
            session.Message(InstallMessage.Error | (InstallMessage)MessageIcon.Error, record);
            record.Close();
        }
    }
}
