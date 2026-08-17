# git-ai-commit

Génère un message de commit au format `Type - Quoi - Ticket` à partir du diff indexé,
via l'API Anthropic. S'utilise comme une sous-commande git : `git ai-commit`.

## Comment ça marche

Git exécute automatiquement tout binaire nommé `git-ai-commit` présent dans le `PATH`
lorsqu'on tape `git ai-commit` — aucun alias à configurer. L'outil :

1. lit le diff indexé (`git diff --cached`) ;
2. extrait le numéro de ticket du nom de branche (ex. `feature/VAL-482-...` → `VAL-482`) ;
3. demande au modèle un message au format `Type - Quoi - Ticket`, avec un corps
   optionnel (puces façon changelog) ;
4. propose de valider, éditer, régénérer ou annuler avant de committer.

## Installation

Prérequis : Go ≥ 1.22.

```bash
# compiler le binaire
go build -o git-ai-commit .

# le déplacer dans un dossier de votre PATH (ex: /usr/local/bin ou ~/.local/bin)
# IMPORTANT : Le binaire DOIT s'appeler exactement 'git-ai-commit'
sudo mv git-ai-commit /usr/local/bin/
```

Si après l'installation, `git ai-commit` n'est pas reconnu :
1. Vérifiez que `/usr/local/bin` est bien dans votre `$PATH`.
2. Essayez de redémarrer votre terminal.
3. Vérifiez les permissions : `chmod +x /usr/local/bin/git-ai-commit`.
4. Tapez `which git-ai-commit` pour confirmer l'emplacement.

Ou via `go install` :
```bash
go install .
# Assurez-vous que $GOPATH/bin (souvent ~/go/bin) est dans votre PATH.
```

## Clé API

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Usage

```bash
git add -p
git ai-commit            # génère, puis propose [v]alider / [e]diter / [r]égénérer / [q]uitter

git ai-commit -a         # git add -A automatique avant de générer
git ai-commit -r         # amende le dernier commit (équivalent git commit --amend)
git ai-commit -y         # valide directement sans confirmation
git ai-commit --print    # affiche le message sans committer
git ai-commit --model claude-sonnet-5
```

Lors de la régénération (`r`), tu peux ajouter une consigne libre
(« insiste sur la migration Doctrine », « type = fix », etc.).

## Configuration

Optionnelle. L'outil lit, dans l'ordre :

1. `.git-ai-commit.json` à la racine du dépôt (réglages par projet) ;
2. `~/.config/git-ai-commit/config.json` (réglages globaux).

Les champs absents gardent leur valeur par défaut. Voir
`.git-ai-commit.example.json`.

| Champ            | Défaut                                                   | Rôle                                        |
|------------------|----------------------------------------------------------|---------------------------------------------|
| `model`          | `claude-haiku-4-5-20251001`                              | Modèle Anthropic                            |
| `language`       | `français`                                               | Langue du message                           |
| `types`          | `feat, fix, refactor, docs, style, test, chore, perf`    | Vocabulaire autorisé pour `Type`            |
| `ticket_pattern` | `[A-Z]{2,}-\d+`                                           | Regex d'extraction du ticket dans la branche|
| `max_diff_chars` | `12000`                                                  | Troncature du diff envoyé au modèle         |
| `generate_body`  | `true`                                                   | Ajoute un corps en puces sous le sujet      |

## Tests

```bash
go test ./...
```
