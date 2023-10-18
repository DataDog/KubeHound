# Frequently Asked Questions

**How long does KubeHound take to run?**

As always.. it depends :) The bulk of the work comes in building up the attack graph. The more edges (i.e attacks) are present, the longer it will take to run. So yet another incentive for more secure clusters! Typical run times we have observed during our testing:

| Cluster Size (Pods) | Duration |
| --------------------|----------|
| 1,000 | <1 min |
| 10,000 | 3 mins |
| 30,000 | 7 mins |

**What happens when you run KubeHound multiple times?**

The data from the previous run is wiped automatically. For very large graphs this can be quite slow and it might be faster to do a hard reset of the backend.