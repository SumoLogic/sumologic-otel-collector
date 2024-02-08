using System.Collections.Generic;
using System.IO;
using YamlDotNet.RepresentationModel;
using YamlDotNet.Serialization;

namespace SumoLogic.wixext
{
    public class ConfigUpdater
    {
        public YamlDocument Document { get; set; }

        public ConfigUpdater(StreamReader streamReader) {
            try
            {
                var ys = new YamlStream();
                ys.Load(streamReader);
                if (ys.Documents.Count == 0)
                {
                    throw new EmptyConfigException("config file is empty");
                }
                this.Document = ys.Documents[0];
            }
            finally
            {
                streamReader.Dispose();
            }
        }

        public void Update(Config config)
        {
            if (config.RemotelyManaged && string.IsNullOrEmpty(config.OpAmpFolder))
            {
                throw new MissingConfigurationException("OpAmpFolder");
            }

            YamlMappingNode root = (YamlMappingNode)this.Document.RootNode;

            EnsureMapKey(root, "extensions");
            YamlMappingNode extensions = (YamlMappingNode)root.Children["extensions"];

            EnsureMapKey(extensions, "sumologic");
            YamlMappingNode sumologic = (YamlMappingNode)extensions.Children["sumologic"];

            if (config.InstallationToken != "")
            {
                EnsureScalarKey(sumologic, "installation_token");
                sumologic.Children["installation_token"] = config.InstallationToken;
            }

            if (config.CollectorFields.Count > 0)
            {
                EnsureMapKey(sumologic, "collector_fields");
                YamlMappingNode collectorFields = (YamlMappingNode)sumologic.Children["collector_fields"];

                foreach (KeyValuePair<string, string> field in config.CollectorFields)
                {
                    EnsureScalarKey(collectorFields, field.Key);
                    collectorFields.Children[field.Key] = field.Value;
                }
            }

            if (config.RemotelyManaged)
            {
                EnsureMapKey(extensions, "opamp");
                YamlMappingNode opamp = (YamlMappingNode)extensions.Children["opamp"];
                EnsureScalarKey(opamp, "remote_configuration_directory");
                opamp.Children["remote_configuration_directory"] = config.OpAmpFolder;

                // Add OpAmp extension to service section
                EnsureMapKey(root, "service");
                YamlMappingNode service = (YamlMappingNode)root.Children["service"];
                EnsureSequenceKey(service, "extensions");
                YamlSequenceNode serviceExtensions = (YamlSequenceNode)service.Children["extensions"];
                if (!serviceExtensions.Children.Contains("opamp"))
                {
                    serviceExtensions.Children.Add("opamp");
                }
            }

            if (config.Ephemeral)
            {
                EnsureScalarKey(sumologic, "ephemeral");
                sumologic.Children["ephemeral"] = "true";
            }

            if (!string.IsNullOrEmpty(config.Api))
            {
                EnsureScalarKey(sumologic, "api_base_url");
                sumologic.Children["api_base_url"] = config.Api;
            }

            // Make sure the sumologic processor node is a map node, otherwise an empty string
            // is generated as the value instead of an empty node.
            if (root.Children.ContainsKey("processors"))
            {
                EnsureMapKey(root, "processors");
                YamlMappingNode processors = (YamlMappingNode)root.Children["processors"];
                if (processors.Children.ContainsKey("sumologic"))
                {
                    EnsureMapKey(processors, "sumologic");
                }
            }
        }

        public void Save(StreamWriter streamWriter)
        {
            try
            {
                var serializer = new Serializer();
                serializer.Serialize(streamWriter, this.Document.RootNode);
            }
            finally
            {
                streamWriter.Flush();
            }
        }

        private void EnsureMapKey(YamlMappingNode node, string key)
        {
            if (node.Children.ContainsKey(key))
            {
                if (node.Children[key].NodeType == YamlNodeType.Mapping) {
                    return;
                }

                // TODO: is this how we want to handle incorrect node types?
                // YamlNode is wrong type, remove it
                node.Children.Remove(key);
            }
            // Add empty YamlMappingNode to key
            node.Children.Add(key, new YamlMappingNode());
        }

        private void EnsureScalarKey(YamlMappingNode node, string key)
        {
            if (node.Children.ContainsKey(key))
            {
                if (node.Children[key].NodeType == YamlNodeType.Scalar)
                {
                    return;
                }

                // TODO: is this how we want to handle incorrect node types?
                // YamlNode is wrong type, remove it
                node.Children.Remove(key);
            }
            // Add empty YamlScalarNode to key
            node.Children.Add(key, new YamlScalarNode());
        }

        private void EnsureSequenceKey(YamlMappingNode node, string key)
        {
            if (node.Children.ContainsKey(key))
            {
                if (node.Children[key].NodeType == YamlNodeType.Sequence)
                {
                    return;
                }

                // TODO: is this how we want to handle incorrect node types?
                // YamlNode is wrong type, remove it
                node.Children.Remove(key);
            }
            // Add empty YamlScalarNode to key
            node.Children.Add(key, new YamlSequenceNode());
        }
    }
}
