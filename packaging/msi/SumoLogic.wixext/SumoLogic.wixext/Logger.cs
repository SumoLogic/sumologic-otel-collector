using WixToolset.Dtf.WindowsInstaller;

namespace SumoLogic.wixext
{
    class Logger
    {
        private string actionName;
        private Session session;

        public Logger(string actionName, Session session)
        {
            this.actionName = actionName;
            this.session = session;
        }

        public void Log(string msg)
        {
            this.session.Log(this.Format(msg));
        }
        public void Log(string format, params object[] args)
        {
            this.session.Log(this.Format(format), args);
        }

        private string Format(string msg)
        {
            return string.Format("{0}: {1}", this.actionName, msg);
        }
    }
}
