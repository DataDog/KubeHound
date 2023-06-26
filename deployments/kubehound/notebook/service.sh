source /tmp/venv/bin/activate
cd "${WORKING_DIR}"
if [ ${GRAPH_NOTEBOOK_SSL} = "" ]; then
    GRAPH_NOTEBOOK_SSL="True"
fi

if [ ${PROVIDE_EXAMPLES} -eq 1 ]; then
    python3 -m graph_notebook.notebooks.install --destination "${EXAMPLE_NOTEBOOK_DIR}"
fi

python3 -m graph_notebook.configuration.generate_config \
    --host "${GRAPH_NOTEBOOK_HOST}" \
    --port "${GRAPH_NOTEBOOK_PORT}" \
    --proxy_host "${GRAPH_NOTEBOOK_HOST}" \
    --proxy_port "${GRAPH_NOTEBOOK_PORT}" \
    --auth_mode "${GRAPH_NOTEBOOK_AUTH_MODE}" \
    --ssl "${GRAPH_NOTEBOOK_SSL}" \
    --iam_credentials_provider "${GRAPH_NOTEBOOK_IAM_PROVIDER}" \
    --load_from_s3_arn "${NEPTUNE_LOAD_FROM_S3_ROLE_ARN}" \
    --aws_region "${AWS_REGION}" \



##### Running The Notebook Service #####
mkdir ~/.jupyter
if [ ! ${NOTEBOOK_PASSWORD} ];
    then
        echo "c.NotebookApp.password='$(python -c "from notebook.auth import passwd; print(passwd('`curl -s 169.254.169.254/latest/meta-data/instance-id`'))")'" >> ~/.jupyter/jupyter_notebook_config.py
else
    echo "c.NotebookApp.password='$(python -c "from notebook.auth import passwd; print(passwd('${NOTEBOOK_PASSWORD}'))")'" >> ~/.jupyter/jupyter_notebook_config.py
fi
echo "c.NotebookApp.allow_remote_access = True" >> ~/.jupyter/jupyter_notebook_config.py
echo "c.InteractiveShellApp.extensions = ['graph_notebook.magics']" >> ~/.jupyter/jupyter_notebook_config.py

nohup jupyter notebook --ip='*' --port ${NOTEBOOK_PORT} "${WORKING_DIR}/notebooks" --allow-root > jupyterserver.log &
nohup jupyter lab --ip='*' --port ${LAB_PORT} "${WORKING_DIR}/notebooks" --allow-root > jupyterlab.log &
tail -f /dev/null