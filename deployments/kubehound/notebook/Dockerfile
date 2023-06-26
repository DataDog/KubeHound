# This Dockerfile is a tailored version of https://github.com/aws/graph-notebook under APACHE 2 LICENCE

FROM amazonlinux:2

# Notebook Port
EXPOSE 8888
# Lab Port
EXPOSE 8889
USER root

# May need to be set to `pipargs=' -i https://pypi.tuna.tsinghua.edu.cn/simple '` for china regions
ENV pipargs=""
ENV WORKING_DIR="/root"
ENV NOTEBOOK_DIR="${WORKING_DIR}/notebooks"
ENV EXAMPLE_NOTEBOOK_DIR="${NOTEBOOK_DIR}/Example Notebooks"
ENV NODE_VERSION=14.x
ENV GRAPH_NOTEBOOK_AUTH_MODE="DEFAULT"
ENV GRAPH_NOTEBOOK_HOST="neptune.cluster-XXXXXXXXXXXX.us-east-1.neptune.amazonaws.com"
ENV GRAPH_NOTEBOOK_PROXY_PORT="8192"
ENV GRAPH_NOTEBOOK_PROXY_HOST=""
ENV GRAPH_NOTEBOOK_PORT="8182"
ENV NEPTUNE_LOAD_FROM_S3_ROLE_ARN=""
ENV AWS_REGION="us-east-1"
ENV NOTEBOOK_PORT="8888"
ENV LAB_PORT="8889"
ENV GRAPH_NOTEBOOK_SSL="True"
ENV NOTEBOOK_PASSWORD="admin"
ENV PROVIDE_EXAMPLES=1


# "when the SIGTERM signal is sent to the docker process, it immediately quits and all established connections are closed"
# "graceful stop is triggered when the SIGUSR1 signal is sent to the docker process"
STOPSIGNAL SIGUSR1


RUN mkdir -p "${WORKING_DIR}" && \
    mkdir -p "${NOTEBOOK_DIR}" && \
    mkdir -p "${EXAMPLE_NOTEBOOK_DIR}" && \
    # Yum Update and install dependencies
    yum update -y && \
    yum install tar gzip git amazon-linux-extras which -y && \
    # Install NPM/Node
    curl --silent --location https://rpm.nodesource.com/setup_${NODE_VERSION} | bash - && \
    yum install nodejs -y && \
    npm install -g opencollective && \
    # Install Python 3.8
    amazon-linux-extras install python3.8 -y && \
    update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.8 1 && \
    echo 'Using python version:' && \
    python3 --version && \
    python3 -m ensurepip --upgrade  && \
    python3 -m venv /tmp/venv && \
    source /tmp/venv/bin/activate && \
    cd "${WORKING_DIR}" && \
    # Clone the repo and install python dependencies
    git clone https://github.com/aws/graph-notebook && \
    cd "${WORKING_DIR}/graph-notebook" && \
    pip3 install --upgrade pip setuptools wheel && \
    pip3 install twine==3.7.1 && \
    pip3 install -r requirements.txt && \
    pip3 install "jupyterlab>=3,<4" && \
    # Build the package
    python3 setup.py sdist bdist_wheel && \
    # install the copied repo
    pip3 install . && \
    # copy premade starter notebooks
    cd "${WORKING_DIR}/graph-notebook" && \
    python3 -m graph_notebook.notebooks.install --destination "${EXAMPLE_NOTEBOOK_DIR}" && \
    jupyter nbextension enable  --py --sys-prefix graph_notebook.widgets && \
    # This allows for the `.ipython` to be set
    python -m graph_notebook.start_jupyterlab --jupyter-dir "${NOTEBOOK_DIR}" && \
    # Cleanup
    yum clean all && \
    yum remove wget tar git  -y && \
    rm -rf /var/cache/yum && \
    rm -rf "${WORKING_DIR}/graph-notebook" && \
    rm -rf /root/.cache && \
    rm -rf /root/.npm/_cacache && \
    rm -rf /usr/share

ADD "KubeHound.ipynb" "${NOTEBOOK_DIR}/KubeHound.ipynb"
ADD ./service.sh /usr/bin/service.sh
RUN chmod +x /usr/bin/service.sh

ENTRYPOINT [ "bash","-c","service.sh" ]
