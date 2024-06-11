name: kubehound-release
services:
  mongodb:
    ports:
      - "127.0.0.1:27017:27017"
    volumes:
      - mongodb_data:/data/db

  kubegraph:
    image: ghcr.io/datadog/kubehound-graph:{{ .VersionTag }}
    ports:
      - "127.0.0.1:8182:8182"
      - "127.0.0.1:8099:8099"
    volumes:
      - kubegraph_data:/var/lib/janusgraph
  
  ui:
    image: ghcr.io/datadog/kubehound-ui:{{ .VersionTag }}
    restart: unless-stopped
    ports:
      - "127.0.0.1:8888:8888"
    networks:
      - kubenet
    labels:
      com.datadoghq.ad.logs: '[{"app": "kubeui", "service": "kubehound"}]'
    volumes:
      - kubeui_data:/root/notebooks/shared

volumes:
  mongodb_data:
  kubegraph_data:
  kubeui_data:

networks:
  kubenet: