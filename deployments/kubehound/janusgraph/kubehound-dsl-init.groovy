import com.datadog.ase.kubehound.KubeHoundTraversalSource

// an init script that returns a Map allows explicit setting of global bindings.
def globals = [:]

// define the default TraversalSource to bind queries to - this one will be named "g".
globals << [kh : traversal(KubeHoundTraversalSource.class).withEmbedded(graph)]