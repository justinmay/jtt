# jtt
Justin's Transcription Tool. A shitty tool to do voice to text for llms running locally. 

## usage
1. hold [ and ] at the same time, let go when done talking. it will paste at your cursor when its done processing
2. run `jtt start` and `jtt stop`. transcription will be pasted to clipboard

## Prereqs  
1. install whisper `brew install whisper-cpp` (the inference engine)
2. download whisper model (the actual weights):
   ```bash
   mkdir -p ~/.local/share/jtt
   curl -L -o ~/.local/share/jtt/ggml-small.en.bin https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-small.en.bin
   ```
   (or use `ggml-base.en.bin` for faster speed, `ggml-tiny.en.bin` for fastest)
3. install sox `brew install sox` (command-line audio lib)
4. install ollama `brew install ollama` (npm / pip for models)
4.5 get ollama runnin in the bg `brew services start ollama` (using homebrew to manage lifecycle)
5. install a model `ollama pull llama3.2:3b` (i'm using a smaller model)
6. install a hotkey lib / write your own `brew install --cask hammerspoon`
7. open hammerspoon and give it permissions for accessibility 
8. double click the .spoon file to install 
9. add the following the the config `hs.loadSpoon("JTT")` and reload the hammerspoon config
10. enjoy

## Install
`pipx install . --force` 


## TODOS
- add a menu bar status icon / settings page (tauri / wails maybe? )
- bug where bracket shows up
- ability to disable ollama (just use whisper). latency is kinda bad right now. 
- config file to point at different models 
- remove hammerspoon dependency 
