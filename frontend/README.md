# TaskFlow Frontend

SvelteKit + Tailwind UI for TaskFlow. It talks to the Task Service (port 8080) and Audit Service (port 8081).

## Environment

Create a `.env` file (or copy `.env.example`) with:

```sh
PUBLIC_TASK_SERVICE_URL=http://localhost:8080
PUBLIC_AUDIT_SERVICE_URL=http://localhost:8081
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.
