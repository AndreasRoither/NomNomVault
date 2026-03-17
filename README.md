<a name="readme-top"></a>

<br />
<div align="center">
  <img src="docs/image/banner.jpg" alt="NomNomVault Banner (placeholder for now)" style="max-width: 800px; width: 100%;" />
  <h1>NomNomVault</h1>

  <p align="center">
    A self-hostable recipe management platform with pantry tracking, meal planning, and smart grocery lists.
    <br />
    <br />
  </p>

  <p>
    <img src="https://img.shields.io/badge/status-In%20Development-yellow" alt="Status: In Development" />
  </p>
</div>

## About The Project

Some recipes are irreplaceable. Your grandmother's apple pie, the soup your mum made when you were sick, that one dish your friend brought to your home party. Like me, you are probably searching for an easy way to save them, not leave them in the tons of cookbooks, scattered screenshots, sticky notes you won't ever touch again.

NomNomVault lets you gather all your recipes in one place and keep them forever. Import from websites, scan old recipe cards, type them in by hand. Once they're here, they're yours. No subscription fees, no ads, no data mining. Just your recipes, stored securely on your own server.

Beyond just storing these recipes, NomNomVault helps you actually use them. Track what's in your pantry so you know what you can cook tonight. Plan your meals for the week. Generate shopping lists that account for what you already have.

The focus of this project is on **self-hosting**. You run it on your own hardware, whether that's a Raspberry Pi, an old laptop, or a NAS.

## Key Features

## Quick Start

### Prerequisites
- `just`
- Docker with `docker compose`
- Go
- Node.js with `pnpm`

### Frontend Only

```bash
just install
just dev
```

- Recipes app: `http://localhost:3000`
- Grocery app: `http://localhost:3001`

### Full Stack

```bash
just install
just compose-up
```

- Recipes app: `http://recipes.localhost`
- Grocery app: `http://grocery.localhost`
- API health check: `http://api.localhost/api/v1/healthz`

### Stop

```bash
just compose-down
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Backend

### Basics

- Backend code lives in `backend/`.
- The API listens on `http://localhost:8080`.
- The worker health endpoint listens on `http://localhost:8081/healthz`.
- PostgreSQL is provided by `just infra-up` or `just compose-up`.

### Startup

Start only the backend dependencies:

```bash
just install
just infra-up
just db-apply
```

Optional: copy the backend env template if you want to override the local defaults:

```bash
cp backend/.env.example backend/.env
```

Run the API locally:

```bash
just backend-dev
```

`just backend-dev` loads `.env` and `backend/.env` when present, and falls back to the local PostgreSQL URL used elsewhere in the repo.

Run the worker locally when needed:

```bash
cd backend
go run ./cmd/worker
```

## Frontend

### Basics

- Frontend apps live in `frontend/apps/recipes-web` and `frontend/apps/grocery-web`.
- Shared frontend API types live in `frontend/packages/api-client`.
- The recipes app runs on `http://localhost:3000`.
- The grocery app runs on `http://localhost:3001`.

### Startup

Start both apps:

```bash
just install
just dev
```

Start one app only:

- `just recipes-dev`
- `just grocery-dev`

Use `just compose-up` if you want the frontend behind Caddy together with the backend and database.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Testing

### Backend Tests
- `just backend-test`

### End-to-End Tests
- `just frontend-test`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Roadmap

- [x] Foundation and developer bootstrap
- [ ] Auth and household foundation
  - Local email and password sign-in
  - Default household bootstrap
  - Household-aware session and route protection
- [ ] Core recipe vault
  - Recipe create, edit, delete, and detail views
  - Ingredients, steps, tags, and recipe media
  - Search and filtering for stored recipes
- [ ] Recipe import workflows
  - URL import into reviewable drafts
  - Raw text import into reviewable drafts
  - OCR pipeline for scanned recipes and handwritten cards
- [ ] Pantry and planning
  - Pantry tracking with expiry awareness
  - Weekly meal planning
    - Preview version will probably not have a calendar view
- [ ] Grocery workflow and offline support
  - Grocery list generation from meal plans and pantry stock
  - Mobile-friendly grocery checklist
  - Offline check-off and viewing for shopping trips
- [ ] Household collaboration and release hardening
  - Invites, roles, and audit-safe household collaboration
  - Notifications, exports, retention, and restore validation
  - Security, observability, and release readiness
- [ ] Post-v1 expansion
  - Standalone NomNomGrocery
  - Cookbook and export improvements
  - OAuth/OIDC and other deferred platform features

But details are TBD.

### First Usable Release Target

The first usable release is complete when a household can sign in, store recipes, import recipes, plan meals, generate grocery lists, and use the grocery checklist on mobile. Once that is done, I will be pushing Docker images for consumption, of course you can also try this out by building it yourself :)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Content Attribution

When importing recipes from external websites, please respect the original content creators' rights. NomNomVault stores the source URL and capture date for attribution purposes. Users are responsible for ensuring they have permission to import and store third-party content.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Built With

TBD


<div align="center">
  Built With ❤️ and Tea 🍵
</div>
