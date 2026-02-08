# Wails Lit Starter ChatBot (WLS ChatBot)

A desktop chat application with AI provider support and model catalog management.

## CLI Commands

### Run Application (GUI)
```bash
wls-chatbot
wls-chatbot --db-path <path>
wls-chatbot --log-level debug|info|warn|error
```

### Generate
```bash
wls-chatbot generate image \
  --provider <provider> \
  --prompt "<description>" \
  [--db-path <path>] \
  [--model <model>] \
  --output <path.png>

wls-chatbot generate image-edit \
  --provider <provider> \
  --prompt "<changes>" \
  --image <input.png> \
  [--db-path <path>] \
  [--mask <mask.png>] \
  [--model <model>] \
  --output <result.png>
```

### Model Management (`model`, alias: `models`)
```bash
wls-chatbot model list \
  [--db-path <path>] \
  [--source seed|user|discovered] \
  [--requires-input-modality <modality>] \
  [--requires-output-modality <modality>] \
  [--requires-capability <capability-id>] \
  [--requires-system-tag <tag>]

wls-chatbot model import [--db-path <path>] --file <custom-models.yaml>
wls-chatbot model sync [--db-path <path>]
```

### Provider Management
```bash
wls-chatbot provider list [--db-path <path>]
wls-chatbot provider test --name <provider> [--db-path <path>]
```

## Custom Models

Place custom model definitions in `<os-user-config-dir>/wls-chatbot/custom-models.yaml` (same structure as `pkg/models/models.yaml`).

Examples:
- Linux: `~/.config/wls-chatbot/custom-models.yaml`
- macOS: `~/Library/Application Support/wls-chatbot/custom-models.yaml`
- Windows: `%AppData%\\wls-chatbot\\custom-models.yaml`

Models are tracked by source:
- **seed** - Bundled with the application
- **user** - Added via CLI import/sync
- **discovered** - Auto-discovered from provider APIs
