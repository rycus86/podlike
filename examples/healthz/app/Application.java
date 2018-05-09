import javax.management.JMException;
import javax.management.MBeanServer;
import javax.management.ObjectName;
import java.io.FileWriter;
import java.io.IOException;
import java.lang.management.ManagementFactory;
import java.util.Date;
import java.util.concurrent.TimeUnit;

public class Application {

    static class Producer {
        final String target;
        final Status stats = new Status();

        Producer(final String target) {
            this.target = target;
        }

        void produce() {
            while (true) {
                writeReport();

                try {
                    Thread.sleep(TimeUnit.SECONDS.toMillis(1));
                } catch (InterruptedException ex) {
                    return;
                }
            }
        }

        void writeReport() {
            final long timestamp = System.currentTimeMillis();

            try (final FileWriter writer = new FileWriter(target)) {
                writer.write("Written at " + new Date(timestamp));
            } catch (IOException ex) {
                ex.printStackTrace();
            }

            stats.lastUpdate = timestamp;
        }

        void start() {
            stats.register();

            final Thread thread = new Thread(this::produce, "ProducerThread");
            thread.setDaemon(false);
            thread.start();
        }
    }

    public interface StatusMBean {
        boolean isAlive();
    }

    public static class Status implements StatusMBean {

        volatile long lastUpdate;

        @Override
        public boolean isAlive() {
            return System.currentTimeMillis() - lastUpdate < TimeUnit.SECONDS.toMillis(5);
        }

        void register() {
            try {
                final MBeanServer mBeanServer = ManagementFactory.getPlatformMBeanServer();
                final ObjectName name = new ObjectName("com.rycus86.example:type=Status");

                mBeanServer.registerMBean(this, name);
            } catch (JMException ex) {
                ex.printStackTrace();
            }
        }
    }

    public static void main(String[] args) {
        final String target = System.getenv().getOrDefault("OUTPUT_TARGET", "/tmp/progress.txt");
        new Producer(target).start();
    }

}
