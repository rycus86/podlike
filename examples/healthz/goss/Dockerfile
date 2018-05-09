FROM aelsabbahy/goss:onbuild

HEALTHCHECK --interval=3s --timeout=3s \
    CMD [ "goss", "-g", "/goss/goss.yaml", "validate" ]

CMD ["serve"]
