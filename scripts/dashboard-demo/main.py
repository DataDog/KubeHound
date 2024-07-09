import pandas as pd
from math import pi

import panel as pn
from gremlin_python.driver.client import Client
from bokeh.palettes import Iridescent, Iridescent
from bokeh.plotting import figure
from bokeh.transform import cumsum
import nest_asyncio

nest_asyncio.apply() # RuntimeError: Cannot run the event loop while another loop is running

GREMLIN_SOCKET = "127.0.0.1:8182"

class KPI:
    DISPLAY_TITLE = ""
    KH_QUERY_MIN_HOPS = ""
    KH_QUERY_EXTERNAL_COUNTS = ""
    KH_QUERY_DETAILS= ''
    KH_QUERY_EXTERNAL_CRITICAL_PATH = ""

    def __init__(self, client):
        self.c = client
        self.res_query_count = c.submit(self.KH_QUERY_EXTERNAL_COUNTS).all().result()
        self.res_query_min_hops = c.submit(self.KH_QUERY_MIN_HOPS).all().result()
        self.res_query_critical_path = c.submit(self.KH_QUERY_EXTERNAL_CRITICAL_PATH).all().result()
        self.get_details()
        print("Loading " + self.DISPLAY_TITLE + " DONE")
    
    def get_main(self):
        return self.DESCRIPTION

    def get_details(self):
        if self.KH_QUERY_DETAILS != '':
            RES_KH_QUERY_VOLUME_INFO = c.submit(self.KH_QUERY_DETAILS).all().result()
            results = []
            for result in c.submit(self.KH_QUERY_DETAILS).all().result():
                data = {}
                for key  in result:
                    data[key] = result[key][0]
                results.append(data)

            self.details_df = pd.DataFrame.from_records(results)

    def display(self):
        return pn.Column(
            f'# {self.DISPLAY_TITLE}', 
            pn.layout.Divider(),
            pn.Row(
                self.DESCRIPTION,
                pn.indicators.Number(
                    value=self.res_query_count[0], name="Total", format="{value:,.0f}", styles=styles
                ),
                pn.indicators.Number(
                    value=self.res_query_min_hops[0], name="Min Hops", format="{value:,.0f}", styles=styles
                ),
                pn.indicators.Number(
                    name='Critical Path', value=int(self.res_query_critical_path[0]), format='{value} %',
                    colors=[(33, 'green'), (66, 'gold'), (100, 'red')]
                ),
            ),
            pn.Row(
                pn.pane.DataFrame(self.details_df, escape=False, index=False, min_height=400, sizing_mode="stretch_both", max_height=250),
            ),
            styles=dict(background='whitesmoke'),
        )

class EndpointKPI(KPI):
    DISPLAY_TITLE = "Endpoints"
    KH_QUERY_MIN_HOPS = "kh.endpoints().minHopsToCritical()"
    KH_QUERY_PORTS = 'kh.endpoints().values("port").dedup()'
    KH_QUERY_EXTERNAL_COUNTS = "kh.endpoints().count()"
    KH_QUERY_DETAILS= 'kh.endpoints().criticalPaths().limit(local,1).dedup().valueMap("serviceEndpoint","port", "namespace")'
    KH_QUERY_EXTERNAL_CRITICAL_PATH = '''kh.V().
    hasLabel("Endpoint").
    count().
    aggregate("t").
    V().
    hasLabel("Endpoint").
    hasCriticalPath().
    count().
    as("e").
    math("100 * e/t").by().by(unfold())'''
    DESCRIPTION = '''
### Exposed asset analyis

The most likely entry points for an attacker into a Kubernetes cluster are:
+ Exposed services via 0day, n-day, or misconfigurations

We can use KubeHound to evaluate the percentage of endpoints/services that can lead to a critical asset.'''

class IdentitiesKPI(KPI):
    DISPLAY_TITLE = "Identities"
    KH_QUERY_MIN_HOPS = "kh.identities().minHopsToCritical()"
    KH_QUERY_EXTERNAL_COUNTS = "kh.identities().count()"
    KH_QUERY_DETAILS= 'kh.identities().criticalPaths().limit(local,1).dedup().valueMap("name","type","namespace")'
    KH_QUERY_EXTERNAL_CRITICAL_PATH = '''kh.V().
    hasLabel("Identity").
    has("critical", false).
    count().
    aggregate("t").
    V().
    hasLabel("Identity").
    has("critical", false).
    hasCriticalPath().
    count().
    as("e").
    math("100 * e/t").by().by(unfold())'''
    DESCRIPTION = '''
### RBAC issues

The most likely entry points for an attacker into a Kubernetes cluster are:
+ Leaked credentials

We can use KubeHound to evaluate the percentage of identities that can lead to a critical asset.'''


class ContainersKPI(KPI):
    DISPLAY_TITLE = "Containers"
    KH_QUERY_MIN_HOPS = "kh.containers().minHopsToCritical()"
    KH_QUERY_EXTERNAL_COUNTS = "kh.containers().count()"
    KH_QUERY_DETAILS= 'kh.containers().criticalPaths().limit(local,1).dedup().valueMap("name","image","app","namespace")'
    KH_QUERY_EXTERNAL_CRITICAL_PATH = '''kh.V().
    hasLabel("Container").
    count().
    aggregate("t").
    V().
    hasLabel("Container").
    hasCriticalPath().
    count().
    as("e").
    math("100 * e/t").by().by(unfold())'''
    DESCRIPTION = '''
### Supply chain attacks

The most likely entry points for an attacker into a Kubernetes cluster are:
+ Supply chain attacks leading to execution within a container

We can use KubeHound to evaluate the percentage of potential compromise containers that can lead to a critical asset.'''


class VolumesKPI(KPI):
    DISPLAY_TITLE = "Volumes"
    KH_QUERY_MIN_HOPS = "kh.volumes().minHopsToCritical()"
    KH_QUERY_EXTERNAL_COUNTS = "kh.volumes().count()"
    KH_QUERY_DETAILS= 'kh.volumes().criticalPaths().limit(local,1).dedup().valueMap("name","sourcePath", "namespace")'
    KH_QUERY_DETAILS_KEYS = ["name", "sourcePath"]
    KH_QUERY_EXTERNAL_CRITICAL_PATH = '''kh.V().
    hasLabel("Volume").
    count().
    aggregate("t").
    V().
    hasLabel("Volume").
    hasCriticalPath().
    count().
    as("e").
    math("100 * e/t").by().by(unfold())'''
    DESCRIPTION = '''
### Volumes issues

The most likely entry points for an attacker into a Kubernetes cluster are:
+ Leaked credentials

We can use KubeHound to evaluate the percentage of potential compromise containers that can lead to a critical asset.'''

class GlobalKPI(KPI):
    DISPLAY_TITLE = "Global stats"
    KH_QUERY_MIN_HOPS = "kh.V().minHopsToCritical()"
    KH_QUERY_EXTERNAL_COUNTS = "kh.V().count()"
    KH_QUERY_DETAILS= ''
    KH_QUERY_DETAILS_KEYS = []
    KH_QUERY_EXTERNAL_CRITICAL_PATH = '''kh.V().
    count().
    aggregate("t").
    V().
    hasCriticalPath().
    count().
    as("e").
    math("100 * e/t").by().by(unfold())'''
    DESCRIPTION = ""

    def get_main(self):
        kh_query_attacks_stats = 'g.E().groupCount().by(label)'
        res_attacks_stats = c.submit(kh_query_attacks_stats).all().result()
        return BokehPieChart(res_attacks_stats[0])

    def display(self):
        return pn.Column(
            f'# {self.DISPLAY_TITLE}', 
            pn.layout.Divider(),
            pn.Row(
                self.get_main(),
                # pn.Spacer(height=200, styles={'background': 'green', 'flex': '3 1 auto'}),
                pn.Column(
                    pn.indicators.Number(
                        value=self.res_query_count[0], name="Total number of k8s entities", format="{value:,.0f}", styles=styles
                    ),
                    pn.indicators.Number(
                        value=self.res_query_min_hops[0], name="Min Hops to cluster admin", format="{value:,.0f}", styles=styles
                    ),
                    pn.indicators.Number(
                        name='Critical Path from any k8s entities', value=int(self.res_query_critical_path[0]), format='{value} %',
                        colors=[(33, 'green'), (66, 'gold'), (100, 'red')]
                    ),
                    margin=(40, 50),
                )
            ),
            styles=dict(background='whitesmoke'),
        )

def BokehPieChart(raw_data):
    data = pd.Series(raw_data).reset_index(name='value').rename(columns={'index':'attacks'})
    data['angle'] = data['value']/data['value'].sum() * 2*pi
    data['color'] = Iridescent[len(raw_data)]

    p = figure( title="Attacks distribution", toolbar_location=None,
            tools="hover", tooltips="@attacks: @value", x_range=(-0.5, 1.0))

    r = p.wedge(x=0, y=1, radius=0.4,
            start_angle=cumsum('angle', include_zero=True), end_angle=cumsum('angle'),
            line_color="white", fill_color='color', legend_field='attacks', source=data)

    p.axis.axis_label=None
    p.axis.visible=False
    p.grid.grid_line_color = None

    return pn.pane.Bokeh(p, theme="dark_minimal")

def BuildDataDashboard(flex_box, kpis):
    for kpi in kpis:
        flex_box.append(kpi.display())

    return flex_box

def GetClusterName():
    KH_QUERY_NODE_CLUSTER="kh.nodes().values('cluster').dedup()"
    res = c.submit(KH_QUERY_NODE_CLUSTER).all().result()
    return res[0]

ACCENT = "teal"

styles = {
    "box-shadow": "rgba(50, 50, 93, 0.25) 0px 6px 12px -2px, rgba(0, 0, 0, 0.3) 0px 3px 7px -3px",
    "border-radius": "4px",
    "padding": "10px",
}

c = Client(f"ws://{GREMLIN_SOCKET}/gremlin", "kh")

global_kpi = GlobalKPI(c)
endpoints_kpi = EndpointKPI(c)
volumes_kpi = VolumesKPI(c)
containers_kpi = ContainersKPI(c)
identitites_kpi = IdentitiesKPI(c)

pn.extension(design='bootstrap')

pn.config.sizing_mode='stretch_width'
flex_box = pn.FlexBox(pn.pane.PNG("./logo_insomnihack.png", fixed_aspect=True, sizing_mode='scale_width'))
kubehound_logo = pn.pane.PNG("./logo_kubehound.png", sizing_mode='scale_width')

data = BuildDataDashboard(flex_box, [global_kpi, endpoints_kpi, containers_kpi, identitites_kpi, volumes_kpi])
title = f'Security posture for {GetClusterName()}'
pn.template.FastListTemplate(
    title=title,
    sidebar=[kubehound_logo],
    main=[pn.Column(data, sizing_mode="stretch_both")],
    main_layout=None,
    accent=ACCENT,
).servable()