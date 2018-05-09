FROM openjdk:8-jdk-alpine as builder

ADD Application.java /var/app/

WORKDIR /var/app

RUN javac Application.java

# the final image

FROM openjdk:8-jre-alpine

COPY --from=builder /var/app/* /var/app/

WORKDIR /var/app

CMD [ "/usr/bin/java",                                      \
      "-Dcom.sun.management.jmxremote.ssl=false",           \
      "-Dcom.sun.management.jmxremote.authenticate=false",  \
      "-Dcom.sun.management.jmxremote.port=5555",           \
      "Application" ]
