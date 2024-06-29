# Common Operations

### Restarting the backend

The backend stack can be restarted by using:
```bash
kubehound backend down
```

These commands will simply reboot backend services, but persist the data via docker volumes.

### Wiping the database

The backend data can be wiped by using:

```bash
kubehound backend wipe
```

These commands will reboot backend services and wipe all data.
