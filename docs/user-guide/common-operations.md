# Common Operations

### Restarting the backend

The backend stack can be restarted via provided script commands. If running from source:

```bash
make backend-reset
```

Otherwise if using binary releases:

```bash
kubehound.sh backend-reset
```

These commands will simply reboot backend services, but persist the data via docker volumes.

### Wiping the database

The backend data can be wiped via provided script commands. If running from source:

```bash
make backend-reset-hard
```

Otherwise if using binary releases:

```bash
kubehound.sh backend-reset-hard
```

These commands will reboot backend services and wipe all data.
