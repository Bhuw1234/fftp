#!/bin/sh
# DEparrow Docker Entrypoint
# Wraps bacalhau and replaces branding with DEparrow

# Create wrapper that replaces bacalhau -> deparrow in output
create_wrapper() {
    cat > /usr/local/bin/deparrow << 'WRAPPER'
#!/bin/sh
# DEparrow - Global Virtual Machine
# All bacalhau output is rebranded to DEparrow

# Run bacalhau and replace all branding
/usr/local/bin/bacalhau.real "$@" 2>&1 | sed \
    -e 's/bacalhau/DEparrow/g' \
    -e 's/Bacalhau/DEparrow/g' \
    -e 's/BACALHAU/DEPARROW/g' \
    -e 's/\.bacalhau/\.deparrow/g'
WRAPPER
    chmod +x /usr/local/bin/deparrow
}

# Setup wrapper on first run
if [ ! -f /usr/local/bin/deparrow ]; then
    mv /usr/local/bin/bacalhau /usr/local/bin/bacalhau.real
    create_wrapper
fi

# Show DEparrow banner
if [ "$1" = "serve" ]; then
    echo ""
    echo "  ██████╗ ██████╗ ██████╗ ███████╗██████╗ "
    echo "  ██╔══██╗██╔══██╗██╔══██╗██╔════╝██╔══██╗"
    echo "  ██║  ██║██████╔╝██████╔╝█████╗  ██████╔╝"
    echo "  ██║  ██║██╔═══╝ ██╔═══╝ ██╔══╝  ██╔══██╗"
    echo "  ██████╔╝██║     ██║     ███████╗██║  ██║"
    echo "  ╚═════╝ ╚═╝     ╚═╝     ╚══════╝╚═╝  ╚═╝"
    echo ""
    echo "  DEparrow - Global Virtual Machine"
    echo "  'AI Agents Buy Compute to Run Themselves'"
    echo ""
fi

# Run DEparrow (wrapper)
exec /usr/local/bin/deparrow "$@"
