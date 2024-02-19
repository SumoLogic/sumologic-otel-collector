using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;

namespace SumoLogic.wixext
{
    public class EmptyConfigException : Exception
    {
        public EmptyConfigException(string message) { }
    }

    public class TagsSyntaxException : Exception
    {
        public TagsSyntaxException(string message) { }
    }

    public class TagSyntaxException : Exception
    {
        public TagSyntaxException(string message) { }
    }

    public class TagsLimitExceededException : Exception
    {
        public TagsLimitExceededException(string message) { }
    }

    public class TagKeyLengthExceededException : Exception
    {
        public TagKeyLengthExceededException(string message) { }
    }

    public class TagValueLengthExceededException : Exception
    {
        public TagValueLengthExceededException(string message) { }
    }

    public class MissingConfigurationException : Exception
    {
        public MissingConfigurationException(string message) { }
    }
}
