# This Dockerfile is a tailored version of https://github.com/aws/graph-notebook under APACHE 2 LICENCE

source /tmp/venv/bin/activate
cd "${WORKING_DIR}"

python3 -m graph_notebook.configuration.generate_config \
    --host "${GRAPH_NOTEBOOK_HOST}" \
    --port "${GRAPH_NOTEBOOK_PORT}" \
    --auth_mode "${GRAPH_NOTEBOOK_AUTH_MODE}" \
    --ssl "${GRAPH_NOTEBOOK_SSL}"

python3 -m graph_notebook.ipython_profile.configure_ipython_profile

##### Running The Notebook Service #####
mkdir ~/.jupyter
if [ ! ${NOTEBOOK_PASSWORD} ];
    then
        echo "No password set for notebook"
        exit 1
else
    echo "c.NotebookApp.password='$(python -c "from notebook.auth import passwd; print(passwd('${NOTEBOOK_PASSWORD}'))")'" >> ~/.jupyter/jupyter_notebook_config.py
fi
echo "c.NotebookApp.allow_remote_access = True" >> ~/.jupyter/jupyter_notebook_config.py
echo "c.InteractiveShellApp.extensions = ['graph_notebook.magics']" >> ~/.jupyter/jupyter_notebook_config.py

# adding all presets notebooks to the trusted list to enable auto-run without warnings
jupyter trust $EXAMPLE_NOTEBOOK_DIR/*.ipynb

nohup jupyter notebook --ip='*' --port ${NOTEBOOK_PORT} "${WORKING_DIR}/notebooks" --allow-root > jupyterserver.log &
nohup jupyter lab --ip='*' --port ${LAB_PORT} "${WORKING_DIR}/notebooks" --allow-root > jupyterlab.log &
tail -f /dev/null
