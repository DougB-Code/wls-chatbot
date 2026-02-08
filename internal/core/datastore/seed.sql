-- seed.sql initializes the database with default data.

-- Initial App Config
INSERT OR IGNORE INTO app_config (id, config_json, created_at, updated_at)
VALUES (
    1,
    '{
        "theme": "dark",
        "providers": [
            {
                "name": "openai",
                "displayName": "OpenAI",
                "type": "openai",
                "enabled": true,
                "baseUrl": "https://api.openai.com/v1",
                "defaultModel": "gpt-4o",
                "inputs": {}
            },
            {
                "name": "anthropic",
                "displayName": "Anthropic",
                "type": "anthropic",
                "enabled": true,
                "baseUrl": "https://api.anthropic.com",
                "defaultModel": "claude-3-5-sonnet-20240620",
                "inputs": {}
            },
            {
                "name": "gemini",
                "displayName": "Google Gemini",
                "type": "gemini",
                "enabled": true,
                "baseUrl": "https://generativelanguage.googleapis.com/v1beta",
                "defaultModel": "gemini-1.5-pro",
                "inputs": {}
            },
            {
                "name": "grok",
                "displayName": "Grok (xAI)",
                "type": "grok",
                "enabled": true,
                "baseUrl": "https://api.x.ai/v1",
                "defaultModel": "grok-beta",
                "inputs": {}
            },
             {
                "name": "openrouter",
                "displayName": "OpenRouter",
                "type": "openrouter",
                "enabled": true,
                "baseUrl": "https://openrouter.ai/api/v1",
                "defaultModel": "openai/gpt-4o-mini",
                "inputs": {}
            }
        ]
    }',
    strftime('%s', 'now') * 1000,
    strftime('%s', 'now') * 1000
);

-- Populate Catalog Providers matches App Config for consistency
INSERT OR IGNORE INTO catalog_providers (id, name, display_name, adapter_type, trust_mode, base_url, created_at, updated_at)
VALUES 
('prov_openai', 'openai', 'OpenAI', 'openai', 'user_managed', 'https://api.openai.com/v1', strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000),
('prov_anthropic', 'anthropic', 'Anthropic', 'anthropic', 'user_managed', 'https://api.anthropic.com', strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000),
('prov_gemini', 'gemini', 'Google Gemini', 'gemini', 'user_managed', 'https://generativelanguage.googleapis.com/v1beta', strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000),
('prov_grok', 'grok', 'Grok (xAI)', 'openai', 'user_managed', 'https://api.x.ai/v1', strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000),
('prov_openrouter', 'openrouter', 'OpenRouter', 'openai', 'user_managed', 'https://openrouter.ai/api/v1', strftime('%s', 'now') * 1000, strftime('%s', 'now') * 1000);
