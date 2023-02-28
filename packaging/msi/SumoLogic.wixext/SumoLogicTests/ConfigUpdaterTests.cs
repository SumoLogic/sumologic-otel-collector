using Microsoft.VisualStudio.TestTools.UnitTesting;
using SumoLogic.wixext;
using System;
using System.IO;
using YamlDotNet.RepresentationModel;

namespace SumoLogicTests
{
    [TestClass]
    public class ConfigUpdaterTests
    {
        string testDataPath = Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "TestData");

        public void InstallationTokenAssertions(Config config, StreamReader sr)
        {
            YamlStream ys = new YamlStream();
            ys.Load(sr);
            YamlMappingNode root = (YamlMappingNode)ys.Documents[0].RootNode;

            Assert.IsTrue(root.Children.ContainsKey("extensions"));
            Assert.AreEqual(YamlNodeType.Mapping, root.Children["extensions"].NodeType);
            var extensions = (YamlMappingNode)root.Children["extensions"];

            Assert.IsTrue(extensions.Children.ContainsKey("sumologic"));
            Assert.AreEqual(YamlNodeType.Mapping, extensions.Children["sumologic"].NodeType);
            var sumologic = (YamlMappingNode)extensions.Children["sumologic"];

            Assert.IsTrue(sumologic.Children.ContainsKey("installation_token"));
            Assert.AreEqual(YamlNodeType.Scalar, sumologic.Children["installation_token"].NodeType);
            Assert.AreEqual(config.InstallationToken, sumologic.Children["installation_token"].ToString());

            Assert.IsTrue(sumologic.Children.ContainsKey("collector_fields"));
            Assert.AreEqual(YamlNodeType.Mapping, sumologic.Children["collector_fields"].NodeType);
            var collectorFields = (YamlMappingNode)sumologic.Children["collector_fields"];

            Assert.IsTrue(collectorFields.Children.ContainsKey("foo"));
            Assert.IsTrue(collectorFields.Children.ContainsKey("baz"));
            Assert.IsTrue(collectorFields.Children.ContainsKey("xaz"));
            Assert.AreEqual(YamlNodeType.Scalar, collectorFields.Children["foo"].NodeType);
            Assert.AreEqual(YamlNodeType.Scalar, collectorFields.Children["baz"].NodeType);
            Assert.AreEqual(YamlNodeType.Scalar, collectorFields.Children["xaz"].NodeType);
            Assert.AreEqual("bar", collectorFields.Children["foo"]);
            Assert.AreEqual("kaz", collectorFields.Children["baz"]);
            Assert.AreEqual("yaz", collectorFields.Children["xaz"]);
        }

        [TestMethod]
        public void TestUpdate_WithExtensionsBlock()
        {
            var filePath = Path.Combine(testDataPath, "with-extensions-block.yaml");
            var config = new Config();
            config.InstallationToken = "foobar";
            config.SetCollectorFieldsFromTags(@"foo=bar,baz=kaz,xaz=yaz");

            using (MemoryStream ms = new MemoryStream())
            {
                var configUpdater = new ConfigUpdater(new StreamReader(filePath));
                configUpdater.Update(config);
                configUpdater.Save(new StreamWriter(ms));

                ms.Seek(0, SeekOrigin.Begin);

                StreamReader sr = new StreamReader(ms);
                while (!sr.EndOfStream)
                {
                    Console.WriteLine(sr.ReadLine());
                }

                ms.Seek(0, SeekOrigin.Begin);

                InstallationTokenAssertions(config, sr);
            }
        }

        [TestMethod]
        public void TestUpdate_WithoutExtensionsBlock()
        {
            var filePath = Path.Combine(testDataPath, "without-extensions-block.yaml");
            var config = new Config();
            config.InstallationToken = "foobar";
            config.SetCollectorFieldsFromTags(@"foo=bar,baz=kaz,xaz=yaz");

            using (MemoryStream ms = new MemoryStream())
            {
                var configUpdater = new ConfigUpdater(new StreamReader(filePath));
                configUpdater.Update(config);
                configUpdater.Save(new StreamWriter(ms));

                ms.Seek(0, SeekOrigin.Begin);

                StreamReader sr = new StreamReader(ms);
                while (!sr.EndOfStream)
                {
                    Console.WriteLine(sr.ReadLine());
                }

                ms.Seek(0, SeekOrigin.Begin);

                InstallationTokenAssertions(config, sr);
            }
        }

        [TestMethod]
        public void TestUpdate_NoIndentation()
        {
            var filePath = Path.Combine(testDataPath, "no-indentation.yaml");
            var config = new Config();
            config.InstallationToken = "foobar";
            config.SetCollectorFieldsFromTags(@"foo=bar,baz=kaz,xaz=yaz");

            using (MemoryStream ms = new MemoryStream())
            {
                var configUpdater = new ConfigUpdater(new StreamReader(filePath));
                configUpdater.Update(config);
                configUpdater.Save(new StreamWriter(ms));

                ms.Seek(0, SeekOrigin.Begin);

                StreamReader sr = new StreamReader(ms);
                while (!sr.EndOfStream)
                {
                    Console.WriteLine(sr.ReadLine());
                }

                ms.Seek(0, SeekOrigin.Begin);

                InstallationTokenAssertions(config, sr);
            }
        }

        [TestMethod]
        [ExpectedException(typeof(EmptyConfigException), "Loading an empty config did not throw an EmptyConfigException.")]
        public void TestUpdate_Empty()
        {
            var filePath = Path.Combine(testDataPath, "empty.yaml");
            var config = new Config();
            config.InstallationToken = "foobar";
            config.SetCollectorFieldsFromTags(@"foo=bar,baz=kaz,xaz=yaz");

            using (MemoryStream ms = new MemoryStream())
            {
                var configUpdater = new ConfigUpdater(new StreamReader(filePath));
            }
        }

        [TestMethod]
        [ExpectedException(typeof(YamlDotNet.Core.SyntaxErrorException), "Loading an invalid config did not throw a YamlDotNet.Core.SyntaxErrorException.")]
        public void TestUpdate_Invalid()
        {
            var filePath = Path.Combine(testDataPath, "invalid.yaml");
            var config = new Config();
            config.InstallationToken = "foobar";
            config.SetCollectorFieldsFromTags(@"foo=bar,baz=kaz,xaz=yaz");

            using (MemoryStream ms = new MemoryStream())
            {
                var configUpdater = new ConfigUpdater(new StreamReader(filePath));
            }
        }
    }
}
