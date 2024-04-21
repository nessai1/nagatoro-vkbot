
![cover](cover.png)

# Usage

Before installation, you must have access to OpenAI gpt-4-turbo model and created VK group with chatbot access.

1. Create config

    Copy `config.example.json` to `config.json`; Get Group API Key and confirmation code from group callback settings to config; Get API key from OpenAI API;


2. Build application by `make build` command


3. Run `nagatoro` binary file


# Config description

| Field          | Description                                                                                                                                                                                         |
|----------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| address        | Address of running bot                                                                                                                                                                              |
| preprompts_dir | Path of custom preprompts dir. Dir must contain `init.txt` file, that describes bot behavior                                                                                                        |
| vk             | Object that describes vk chatbot settings. In addition to setting up the connection, there is a field `group_chat_triggers` - this is an array of strings that the bot responds to in group messages|
| openai         | OpenAI API connection settings                                                                                                                                                                      |
