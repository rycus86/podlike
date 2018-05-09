FROM maven:3.5.3-jdk-8-alpine as builder

ADD https://github.com/prometheus/jmx_exporter/archive/parent-0.3.0.zip /tmp/

RUN mkdir /var/exporter && unzip /tmp/parent-0.3.0.zip -d /var/exporter

WORKDIR /var/exporter/jmx_exporter-parent-0.3.0

RUN mvn clean install

# the final image

FROM openjdk:8-jre-alpine

COPY --from=builder \
    /var/exporter/jmx_exporter-parent-0.3.0/jmx_prometheus_httpserver/target/jmx_prometheus_httpserver-0.3.0-jar-with-dependencies.jar \
    /var/app/

ADD exporter.yml /var/conf/

WORKDIR /var/app

ENTRYPOINT [ "/usr/bin/java", "-jar", "jmx_prometheus_httpserver-0.3.0-jar-with-dependencies.jar" ]
CMD [ "5556", "/var/conf/exporter.yml" ]
