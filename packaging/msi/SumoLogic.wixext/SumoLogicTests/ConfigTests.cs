using Microsoft.VisualStudio.TestTools.UnitTesting;
using SumoLogic.wixext;
using System;
using System.Linq;

namespace SumoLogicTests
{
    [TestClass]
    public class ConfigTests
    {
        private static readonly Random random = new Random();

        public static string RandomString(int length)
        {
            const string chars = "abcdefghijklmnopqrstuvwxyz";
            return new string(Enumerable.Repeat(chars, length)
                .Select(s => s[random.Next(s.Length)]).ToArray());
        }

        [TestMethod]
        public void TestSetCollectorFieldsFromTags_Valid()
        {
            var tags = @"foo=bar,baz=kaz , xaz=yaz";

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);

            Assert.AreEqual(config.CollectorFields.Keys.Count, 3);

            Assert.IsTrue(config.CollectorFields.ContainsKey("foo"));
            Assert.IsTrue(config.CollectorFields.ContainsKey("baz"));
            Assert.IsTrue(config.CollectorFields.ContainsKey("xaz"));

            Assert.AreEqual("bar", config.CollectorFields["foo"]);
            Assert.AreEqual("kaz", config.CollectorFields["baz"]);
            Assert.AreEqual("yaz", config.CollectorFields["xaz"]);
        }

        [TestMethod]
        public void TestSetCollectorFieldsFromTags_Empty()
        {
            var tags = @"";

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);

            Assert.AreEqual(config.CollectorFields.Keys.Count, 0);
        }

        [TestMethod]
        [ExpectedException(typeof(TagsSyntaxException), "Invalid tags syntax did not throw a TagsSyntaxException.")]
        public void TestSetCollectorFieldsFromTags_NoMatches()
        {
            var tags = @"key";

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);
        }

        [TestMethod]
        [ExpectedException(typeof(TagsLimitExceededException), "Tag count exceeding 10 did not throw a TagsLimitExceededException.")]
        public void TestSetCollectorFieldsFromTags_LimitExceeded()
        {
            var tags = @"a=b,c=d,e=f,g=h,i=j,k=l,m=n,o=p,q=r,s=t,u=v";

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);
        }

        [TestMethod]
        [ExpectedException(typeof(TagKeyLengthExceededException), "Tag key exceeding 255 characters did not throw a TagKeyLengthExceededException.")]
        public void TestSetCollectorFieldsFromTags_KeyLengthExceeded()
        {
            var tags = string.Format("{0}=value", RandomString(256));

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);
        }

        [TestMethod]
        [ExpectedException(typeof(TagValueLengthExceededException), "Tag value exceeding 200 characters did not throw a TagValueLengthExceededException.")]
        public void TestSetCollectorFieldsFromTags_ValueLengthExceeded()
        {
            var tags = string.Format("key={0}", RandomString(201));

            var config = new Config();
            config.SetCollectorFieldsFromTags(tags);
        }
    }
}
