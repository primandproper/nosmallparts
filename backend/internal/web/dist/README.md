`make web` copies the Vite build of `frontend/` here for the binary to embed.
Everything but this file is ignored; without a build the server answers the
app's routes with a note saying so.
