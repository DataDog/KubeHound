import com.datadog.ase.kubehound.KubeHoundTraversalSource

// An init script that returns a Map allows explicit setting of global bindings.
def globals = [:]

// Define the custom KubeHoundTraversalSource to bind queries to - this one will be named "kh".
globals << [kh : traversal(KubeHoundTraversalSource.class).withEmbedded(graph)]
