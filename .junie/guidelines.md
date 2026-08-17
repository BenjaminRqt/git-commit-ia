# Directives projet — git-ai-commit

Ce fichier décrit le projet à l'agent Junie. Respecte ces conventions pour toute
génération ou modification de code.

## Vue d'ensemble

`git-ai-commit` est un outil en ligne de commande écrit en **Go** qui génère un
message de commit au format `Type - Quoi - Ticket` à partir du diff indexé, via
l'**API Anthropic**. Il s'utilise comme une sous-commande git : `git ai-commit`.

Mécanisme clé : git exécute automatiquement tout binaire nommé `git-ai-commit`
présent dans le `PATH` quand on tape `git ai-commit`. Le nom du binaire est donc
significatif — ne pas le renommer sans raison.

## Stack & contraintes

- **Go ≥ 1.22**, module nommé `git-ai-commit` (voir `go.mod`).
- **Aucune dépendance externe** : bibliothèque standard uniquement. C'est un choix
  délibéré pour une distribution triviale (un seul binaire autonome). Ne pas
  introduire de dépendance tierce sans nécessité forte ; privilégier `net/http`,
  `encoding/json`, `os/exec`, etc.
- Le **mode module doit être actif** (`GO111MODULE=on`) pour que les imports
  `git-ai-commit/internal/...` se résolvent.
- Tous les **textes destinés à l'utilisateur sont en français** (messages, erreurs,
  invites interactives). Conserver cette convention.

## Arborescence

```
git-ai-commit/
├── go.mod                          # module git-ai-commit, go 1.22
├── main.go                         # point d'entrée, flags, boucle interactive, spinner
├── README.md
├── .git-ai-commit.example.json     # exemple de config surchargeable
└── internal/
    ├── ai/
    │   └── ai.go                   # client API Anthropic + construction des prompts
    ├── config/
    │   └── config.go               # struct Config, chargement fichier JSON + env
    └── gitutil/
        ├── gitutil.go              # wrappers git (diff, branche, ticket, commit)
        └── gitutil_test.go         # test de ExtractTicket
```

Un package = un dossier. Ne jamais placer plusieurs packages dans le même
répertoire (Go l'interdit).

## Responsabilités des packages

- **`main`** (`main.go`) : parsing des flags, orchestration, boucle interactive
  (`[v]alider / [e]diter / [r]égénérer / [q]uitter`), spinner. Toute la logique
  d'erreur remonte via `run() error` puis est affichée avec le préfixe `✖ `.
- **`internal/config`** : `Config` + `loadConfig()`. Part des valeurs par défaut
  (`defaultConfig`/`Default`), applique le premier fichier trouvé (les champs
  absents gardent leur défaut via `json.Unmarshal` sur la struct pré-remplie),
  puis lit la clé API depuis l'environnement.
- **`internal/gitutil`** : toutes les interactions git passent par le helper `run`
  (via `os/exec`), qui enrichit l'erreur avec `stderr`. Le commit se fait par
  fichier temporaire (`git commit -F <tmp>`), avec `-e` pour ouvrir l'éditeur.
- **`internal/ai`** : client HTTP Anthropic et construction des prompts système /
  utilisateur. Ne pas logguer la clé API.

## Contrat de format (à préserver)

Le **sujet** du commit suit strictement : `Type - Quoi - Ticket`.

- `Type` : un mot parmi la liste configurable (`feat, fix, refactor, docs, style,
  test, chore, perf` par défaut).
- `Quoi` : description à l'impératif, ~60 caractères, majuscule initiale, sans point final.
- `Ticket` : référence extraite du nom de branche, ou `N/A`.

Si `generate_body` est activé : une ligne vide puis 1 à 5 puces `- ...` façon
changelog. Cette structure est définie dans `buildSystem` / `buildUser` (package
`ai`) — toute évolution du format se fait là.

Le ticket est extrait du nom de branche par regex (`ticket_pattern`, défaut
`[A-Z]{2,}-\d+`) dans `ExtractTicket`. Exemple : `feature/VAL-482-refonte` → `VAL-482`.

## Configuration

Fichiers lus, dans l'ordre (champs absents = valeurs par défaut) :

1. `.git-ai-commit.json` à la racine du dépôt (réglages par projet) ;
2. `~/.config/git-ai-commit/config.json` (réglages globaux).

Champs : `model`, `language`, `types`, `ticket_pattern`, `max_diff_chars`,
`generate_body`. La clé API vient **exclusivement** de la variable
d'environnement `ANTHROPIC_API_KEY` (jamais d'un fichier, jamais en dur).

## Intégration API Anthropic

- Endpoint : `POST https://api.anthropic.com/v1/messages`.
- En-têtes : `content-type: application/json`, `x-api-key: <clé>`,
  `anthropic-version: 2023-06-01`.
- Modèle par défaut : `claude-haiku-4-5-20251001` (rapide et économique, adapté à
  un diff) — surchargeable via config ou `--model`.
- Le diff envoyé est tronqué à `max_diff_chars` (défaut 12000) pour maîtriser le coût.
- La réponse est parsée en extrayant les blocs `content[].text` de type `text`.

## Flags CLI

- `-y` : committer sans confirmation.
- `-a` : `git add -A` avant de générer.
- `--print` : afficher le message sans committer.
- `--model <id>` : surcharger le modèle.

## Commandes de dev

```bash
go build -o git-ai-commit .     # compiler le binaire (nom exact requis)
go test ./...                   # lancer les tests
go vet ./...                    # analyse statique
gofmt -w .                      # formatage (obligatoire avant commit)
```

Pour tester en conditions réelles comme sous-commande git : placer le binaire
`git-ai-commit` dans un dossier du `PATH`, puis `git ai-commit` dans un dépôt
avec des changements indexés (nécessite `ANTHROPIC_API_KEY`).

## Conventions de code

- Formatage `gofmt` systématique ; nommage Go idiomatique (exporté = PascalCase).
- Les erreurs sont retournées, pas gérées par `panic` ; messages en français,
  contextualisés avec `fmt.Errorf("...: %w", err)`.
- Pas d'effet de bord réseau dans les tests ; le seul test actuel
  (`ExtractTicket`) est pur — garder les tests hors-ligne.
- Ajouter un test lorsqu'on introduit une fonction pure non triviale
  (extraction, parsing, formatage).

## À ne pas casser

- Le nom du binaire `git-ai-commit` (sous-commande git).
- L'absence de dépendances externes.
- Le contrat de format `Type - Quoi - Ticket`.
- La provenance de la clé API (variable d'environnement uniquement).
- Les textes utilisateur en français.