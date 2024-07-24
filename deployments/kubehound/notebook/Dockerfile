# This Dockerfile is a tailored version of https://github.com/aws/graph-notebook under APACHE 2 LICENCE

FROM amazonlinux:2022

# Notebook Port
EXPOSE 8888
# Lab Port
EXPOSE 8889

# May need to be set to `pipargs=' -i https://pypi.tuna.tsinghua.edu.cn/simple '` for china regions
ENV pipargs=""
ENV WORKING_DIR="/kubehound"
ENV NOTEBOOK_DIR="${WORKING_DIR}/notebooks"
ENV EXAMPLE_NOTEBOOK_DIR="${NOTEBOOK_DIR}/kubehound_presets"
ENV NODE_VERSION=14
ENV PYTHON_VERSION=3.10
ENV GRAPH_NOTEBOOK_AUTH_MODE="DEFAULT"
ENV GRAPH_NOTEBOOK_HOST="kubegraph"
ENV GRAPH_NOTEBOOK_PROXY_PORT="8192"
ENV GRAPH_NOTEBOOK_PROXY_HOST=""
ENV GRAPH_NOTEBOOK_PORT="8182"
ENV NEPTUNE_LOAD_FROM_S3_ROLE_ARN=""
ENV AWS_REGION="us-east-1"
ENV NOTEBOOK_PORT="8888"
ENV LAB_PORT="8889"
ENV GRAPH_NOTEBOOK_SSL="True"
ENV NOTEBOOK_PASSWORD="admin"
ENV PROVIDE_EXAMPLES=0

# "when the SIGTERM signal is sent to the docker process, it immediately quits and all established connections are closed"
# "graceful stop is triggered when the SIGUSR1 signal is sent to the docker process"
STOPSIGNAL SIGUSR1

ENV GID 1000
ENV UID 1000

# Update the package list, install sudo, create a non-root user, and grant password-less sudo permissions
RUN yum update -y && \
    yum install -y sudo shadow-utils && \
    /usr/sbin/groupadd --gid $GID nonroot && \
    adduser --uid $UID --gid $GID --system nonroot -m && \
    echo 'nonroot ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

RUN mkdir -p "${WORKING_DIR}" && \
    mkdir -p "${NOTEBOOK_DIR}" && \
    mkdir -p "${EXAMPLE_NOTEBOOK_DIR}" && \
    chown -R nonroot:nonroot "${WORKING_DIR}" && \
    # Yum Update and install dependencies
    yum update -y && \
    yum install tar gzip git findutils -y && \
    # Install NPM/Node
    curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.3/install.sh | bash && \
    . ~/.nvm/nvm.sh && \
    nvm install ${NODE_VERSION} && \
    # Install Python
    yum install python${PYTHON_VERSION} -y && \
    # update-alternatives --install /usr/bin/python3 python3 /usr/bin/python${PYTHON_VERSION} 1 && \
    echo 'Using python version:' && \
    python${PYTHON_VERSION} --version && \
    python${PYTHON_VERSION} -m ensurepip --upgrade  && \
    python${PYTHON_VERSION} -m venv /tmp/venv && \
    source /tmp/venv/bin/activate && \
    cd "${WORKING_DIR}" && \
    # Clone the repo and install python dependencies
    git clone https://github.com/aws/graph-notebook && \
    cd "${WORKING_DIR}/graph-notebook" && \
    chown -R nonroot:nonroot "${WORKING_DIR}/graph-notebook" && \
    pip3 install --upgrade pip setuptools wheel && \
    pip3 install twine==3.7.1 && \
    pip3 install -r requirements.txt && \
    pip3 install --upgrade 'jupyter-server<2.0.0' && \
    pip3 install "jupyterlab>=3,<4" && \
    pip3 install jupyter_contrib_nbextensions && \
    pip3 install jupyter_nbextensions_configurator && \
    # Build the package
    python3 setup.py sdist bdist_wheel && \
    # install the copied repo
    pip3 install . && \
    # copy premade starter notebooks
    cd "${WORKING_DIR}/graph-notebook" && \
    jupyter contrib nbextension install --system --debug && \
    jupyter nbextension enable  --py --sys-prefix graph_notebook.widgets && \
    jupyter nbextensions_configurator enable --system  # can be skipped for notebook >=5.3 && \
    # This allows for the `.ipython` to be set
    python -m graph_notebook.start_jupyterlab --jupyter-dir "${NOTEBOOK_DIR}" && \
    deactivate && \
    # Cleanup
    yum clean all && \
    yum remove wget tar git  -y && \
    rm -rf /var/cache/yum && \
    rm -rf "${WORKING_DIR}/graph-notebook" && \
    rm -rf /root/.cache && \
    rm -rf /root/.npm/_cacache && \
    cd /usr/share && \
    rm -r $(ls -A | grep -v terminfo)

# Set the non-root user as the default user
USER nonroot

ADD --chown=nonroot:nonroot *.ipynb ${EXAMPLE_NOTEBOOK_DIR}/

# Adding support for init_cell - allow cell to be run on startup of the notebook
# Command not working jupyter nbextension enable --system init_cell && \
ADD --chown=nonroot:nonroot notebook.json /home/nonroot/.jupyter/nbconfig/notebook.json

ADD --chown=nonroot:nonroot ./service.sh /usr/bin/service.sh
RUN chmod +x /usr/bin/service.sh

ENTRYPOINT [ "bash","-c","service.sh" ]
