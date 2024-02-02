using System;
using System.Collections.Generic;
using System.Text.RegularExpressions;

namespace SumoLogic.wixext
{
    public class Config
    {
        public string InstallationToken { get; set; }
        public Dictionary<string, string> CollectorFields { get; set; }
        public bool RemotelyManaged { get; set; }
        public bool Ephemeral { get; set; }
        public string OpAmpFolder { get; set; }
        public string Api { get; set; }

        public Config() {
            this.CollectorFields = new Dictionary<string, string>();
        }

        public void SetCollectorFieldsFromTags(string tags)
        {
            if (tags.Length == 0) { return; }

            var tagsRx = new Regex(@"([^=,]+)=([^\0]+?)(?=,[^,]+=|$)", RegexOptions.Compiled);
            var matches = tagsRx.Matches(tags);

            if (matches.Count == 0)
            {
                throw new TagsSyntaxException("tags were provided with invalid syntax");
            }
            if (matches.Count > 10)
            {
                throw new TagsLimitExceededException("the limit of 10 tags was exceeded");
            }

            foreach (Match match in matches)
            {
                if (match.Groups.Count != 3)
                {
                    Console.WriteLine("Groups: {0}", match.Groups.Count);
                    var msg = string.Format("invalid syntax for tag: {0}", match.Value);
                    throw new TagSyntaxException(msg);
                }
                var key = match.Groups[1].Value.Trim();
                var value = match.Groups[2].Value.Trim();

                if (key.Length > 255)
                {
                    var msg = string.Format("tag key exceeds maximum length of 255: {0}", key);
                    throw new TagKeyLengthExceededException(msg);
                }
                if (value.Length > 200)
                {
                    var msg = string.Format("tag value exceeds maximum length of 200: {0}", value);
                    throw new TagValueLengthExceededException(msg);
                }

                this.CollectorFields.Add(key, value);
            }
        }
    }
}
