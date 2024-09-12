# Wiki

The website [kubehound.io](https://kubehound.io) is being statically generated from [docs](https://github.com/DataDog/KubeHound/tree/main/docs) directory. It uses [mkdocs]() under the hood. To generate [kubehound.io](https://kubehound.io) locally use:

```bash
make local-wiki
```

!!! tip

    All the configuration of the website (url, menu, css, ...) is being made from [mkdocs.yml](https://github.com/DataDog/KubeHound/blob/main/mkdocs.yml) file:

## Push new version

The website will get automatically updated everytime there is changemement in [docs](https://github.com/DataDog/KubeHound/tree/main/docs) directory or the [mkdocs.yml](https://github.com/DataDog/KubeHound/blob/main/mkdocs.yml) file. This is being handled by [docs](https://github.com/DataDog/KubeHound/blob/main/.github/workflows/docs.yml) workflow.

!!! note

    The domain for the wiki is being setup in the [CNAME](https://github.com/DataDog/KubeHound/tree/main/docs/CNAME) file.
