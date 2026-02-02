import { useState, useEffect } from 'react';
import { Events } from '@wailsio/runtime';
import * as JTTService from '../bindings/jtt/jttservice.js';
import './App.css';

function App() {
  const [config, setConfig] = useState(null);
  const [state, setState] = useState('idle');
  const [ollamaModels, setOllamaModels] = useState([]);
  const [ollamaRunning, setOllamaRunning] = useState(false);
  const [deps, setDeps] = useState(null);
  const [whisperModels, setWhisperModels] = useState([]);
  const [downloadedModels, setDownloadedModels] = useState([]);
  const [downloading, setDownloading] = useState(null);
  const [installing, setInstalling] = useState(null);
  const [modelsExpanded, setModelsExpanded] = useState(false);
  const [activeTab, setActiveTab] = useState('settings');
  const [history, setHistory] = useState([]);
  const [defaultPrompt, setDefaultPrompt] = useState('');
  const [microphones, setMicrophones] = useState([]);

  useEffect(() => {
    loadData();
    Events.On('state-change', (newState) => setState(newState));
  }, []);

  const loadData = async () => {
    try {
      const [cfg, appState, models, running, depStatus, whisper, downloaded, hist, defPrompt, mics] = await Promise.all([
        JTTService.GetConfig(),
        JTTService.GetState(),
        JTTService.GetOllamaModels(),
        JTTService.IsOllamaRunning(),
        JTTService.CheckDependencies(),
        JTTService.GetAvailableWhisperModels(),
        JTTService.GetDownloadedModels(),
        JTTService.GetHistory(),
        JTTService.GetDefaultPrompt(),
        JTTService.GetMicrophones(),
      ]);
      setConfig(cfg);
      setState(appState);
      setOllamaModels(models || []);
      setOllamaRunning(running);
      setDeps(depStatus);
      setWhisperModels(whisper || []);
      setDownloadedModels(downloaded || []);
      setHistory(hist || []);
      setDefaultPrompt(defPrompt || '');
      setMicrophones(mics || []);
    } catch (err) {
      console.error('Failed to load data:', err);
    }
  };

  const saveConfig = async (updates) => {
    if (!config) return;
    const newConfig = { ...config, ...updates };
    setConfig(newConfig);
    await JTTService.SaveConfig(newConfig);
  };

  const handleInstall = async (dep) => {
    setInstalling(dep);
    await JTTService.InstallDependency(dep);
    await loadData();
    setInstalling(null);
  };

  const handleDownloadModel = async (model) => {
    setDownloading(model.name);
    await JTTService.DownloadWhisperModel(model.name, model.url);
    await loadData();
    setDownloading(null);
  };

  if (!config || !deps) {
    return <div className="loading">Loading...</div>;
  }

  const missingDeps = !deps.sox || !deps.whisper;

  const formatTime = (timestamp) => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  return (
    <div className="app">
      <header className="header">
        <h1>JTT</h1>
        <div className={`status status-${state}`}>
          {state === 'idle' && 'Ready'}
          {state === 'recording' && 'Recording'}
          {state === 'processing' && 'Processing'}
        </div>
      </header>

      <nav className="tabs">
        <button 
          className={`tab ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          Settings
        </button>
        <button 
          className={`tab ${activeTab === 'history' ? 'active' : ''}`}
          onClick={() => { setActiveTab('history'); loadData(); }}
        >
          History
        </button>
        <button 
          className={`tab ${activeTab === 'prompt' ? 'active' : ''}`}
          onClick={() => setActiveTab('prompt')}
        >
          Prompt
        </button>
      </nav>

      {activeTab === 'settings' && (
        <>
          {missingDeps && (
            <section className="section warning">
              <h2>Missing Dependencies</h2>
              <p>Some required dependencies are not installed:</p>
              <div className="deps-list">
                {!deps.sox && (
                  <div className="dep-item">
                    <span>sox (audio recording)</span>
                    <button 
                      onClick={() => handleInstall('sox')}
                      disabled={installing === 'sox'}
                    >
                      {installing === 'sox' ? 'Installing...' : 'Install'}
                    </button>
                  </div>
                )}
                {!deps.whisper && (
                  <div className="dep-item">
                    <span>whisper-cpp (transcription)</span>
                    <button 
                      onClick={() => handleInstall('whisper')}
                      disabled={installing === 'whisper'}
                    >
                      {installing === 'whisper' ? 'Installing...' : 'Install'}
                    </button>
                  </div>
                )}
              </div>
            </section>
          )}

          <section className="section">
            <h2>Microphone</h2>
            <div className="form-group">
              <label>Input Device</label>
              <select
                value={config.microphone || ''}
                onChange={(e) => saveConfig({ microphone: e.target.value })}
              >
                {microphones.map((m) => (
                  <option key={m.id} value={m.id}>{m.name}</option>
                ))}
              </select>
              <p className="hint">
                Select the microphone to use for recording.
              </p>
            </div>
          </section>

          <section className="section">
            <h2>Whisper Model</h2>
            {downloadedModels.length > 0 && (
              <div className="form-group">
                <label>Current Model</label>
                <select
                  value={config.whisperModel}
                  onChange={(e) => saveConfig({ whisperModel: e.target.value })}
                >
                  {downloadedModels.map((m) => (
                    <option key={m} value={`${process.env.HOME || '~'}/.local/share/jtt/${m}`}>{m}</option>
                  ))}
                </select>
              </div>
            )}
            <div className="form-group">
              <label className="toggle">
                <input
                  type="checkbox"
                  checked={config.filterHallucinations !== false}
                  onChange={(e) => saveConfig({ filterHallucinations: e.target.checked })}
                />
                <span>Filter hallucinations</span>
              </label>
              <p className="hint">
                Filter out common whisper hallucinations like "you" when recording silence.
              </p>
            </div>
            <div className="accordion">
              <button 
                className="accordion-header"
                onClick={() => setModelsExpanded(!modelsExpanded)}
              >
                <span>Available Models</span>
                <span className={`accordion-icon ${modelsExpanded ? 'expanded' : ''}`}>&#9662;</span>
              </button>
              {modelsExpanded && (
                <div className="models-table">
                  <table>
                    <thead>
                      <tr>
                        <th>Model</th>
                        <th>Size</th>
                        <th>Speed</th>
                        <th>Quality</th>
                        <th></th>
                      </tr>
                    </thead>
                    <tbody>
                      {whisperModels.map((m) => {
                        const isDownloaded = downloadedModels.includes(`ggml-${m.name}.bin`);
                        return (
                          <tr key={m.name}>
                            <td>{m.name}</td>
                            <td>{m.size}</td>
                            <td>{m.speed}</td>
                            <td>{m.quality}</td>
                            <td>
                              {isDownloaded ? (
                                <span className="badge">Downloaded</span>
                              ) : (
                                <button
                                  onClick={() => handleDownloadModel(m)}
                                  disabled={downloading !== null}
                                >
                                  {downloading === m.name ? 'Downloading...' : 'Download'}
                                </button>
                              )}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          </section>

          <section className="section">
            <h2>Ollama (Text Cleaning)</h2>
            <div className="form-group">
              <label className="toggle">
                <input
                  type="checkbox"
                  checked={config.useOllama}
                  onChange={(e) => saveConfig({ useOllama: e.target.checked })}
                />
                <span>Enable Ollama text cleaning</span>
              </label>
              <p className="hint">
                When enabled, transcriptions are cleaned up using a local LLM to remove filler words and fix punctuation.
              </p>
            </div>
            
            {config.useOllama && (
              <>
                {!ollamaRunning ? (
                  <div className="warning-inline">
                    Ollama is not running. Start it with: <code>brew services start ollama</code>
                  </div>
                ) : (
                  <div className="form-group">
                    <label>Model</label>
                    <select
                      value={config.ollamaModel}
                      onChange={(e) => saveConfig({ ollamaModel: e.target.value })}
                    >
                      {ollamaModels.map((m) => (
                        <option key={m} value={m}>{m}</option>
                      ))}
                    </select>
                  </div>
                )}
              </>
            )}
          </section>

          <section className="section">
            <h2>Shortcut</h2>
            <p className="hint">
              Hold this shortcut to record. Release to stop and transcribe.
            </p>
            <div className="shortcut-config">
              <div className="modifier-buttons">
                {['cmd', 'ctrl', 'alt', 'shift'].map((mod) => (
                  <button
                    key={mod}
                    className={`modifier-btn ${config.hotkey?.modifiers?.includes(mod) ? 'active' : ''}`}
                    onClick={() => {
                      const current = config.hotkey?.modifiers || [];
                      const newMods = current.includes(mod)
                        ? current.filter(m => m !== mod)
                        : [...current, mod];
                      saveConfig({ hotkey: { ...config.hotkey, modifiers: newMods } });
                    }}
                  >
                    {mod === 'cmd' ? '⌘' : mod === 'ctrl' ? '⌃' : mod === 'alt' ? '⌥' : '⇧'}
                  </button>
                ))}
              </div>
              <span className="shortcut-plus">+</span>
              <select
                className="key-select"
                value={config.hotkey?.keys?.[0] || 'r'}
                onChange={(e) => saveConfig({ hotkey: { ...config.hotkey, keys: [e.target.value] } })}
              >
                {[...'abcdefghijklmnopqrstuvwxyz'].map((k) => (
                  <option key={k} value={k}>{k.toUpperCase()}</option>
                ))}
                {['0','1','2','3','4','5','6','7','8','9'].map((k) => (
                  <option key={k} value={k}>{k}</option>
                ))}
                <option value="space">Space</option>
              </select>
            </div>
            <p className="hint" style={{marginTop: '12px'}}>
              Restart the app to apply shortcut changes.
            </p>
          </section>

          <section className="section">
            <h2>Media Control</h2>
            <div className="form-group">
              <label className="toggle">
                <input
                  type="checkbox"
                  checked={config.pauseMediaOnRecord !== false}
                  onChange={(e) => saveConfig({ pauseMediaOnRecord: e.target.checked })}
                />
                <span>Pause media while recording</span>
              </label>
              <p className="hint">
                Automatically pause music/media when recording starts and resume when done.
                {!deps.nowPlaying && (
                  <span className="hint-warning"> Requires nowplaying-cli - <button className="link-btn" onClick={() => handleInstall('nowplaying')} disabled={installing === 'nowplaying'}>{installing === 'nowplaying' ? 'Installing...' : 'Install'}</button></span>
                )}
              </p>
            </div>
          </section>

          <section className="section">
            <h2>Test</h2>
            <div className="test-buttons">
              <button
                className="btn-primary"
                onClick={() => state === 'idle' 
                  ? JTTService.StartRecording() 
                  : JTTService.StopRecording()
                }
                disabled={state === 'processing'}
              >
                {state === 'idle' && 'Start Recording'}
                {state === 'recording' && 'Stop Recording'}
                {state === 'processing' && 'Processing...'}
              </button>
            </div>
          </section>
        </>
      )}

      {activeTab === 'history' && (
        <section className="section">
          <h2>Transcription History</h2>
          {history.length === 0 ? (
            <p className="hint">No transcriptions yet. Record something to see history.</p>
          ) : (
            <div className="history-list">
              {[...history].reverse().map((entry, idx) => (
                <div key={entry.timestamp} className="history-entry">
                  <div className="history-header">
                    <span className="history-time">{formatTime(entry.timestamp)}</span>
                  </div>
                  <div className="history-row">
                    <div className="history-label">
                      Whisper <span className="history-timing">({entry.whisperTime.toFixed(2)}s)</span>
                    </div>
                    <div className="history-output">{entry.whisperOutput}</div>
                  </div>
                  <div className="history-row">
                    <div className="history-label">
                      LLM <span className="history-timing">({entry.llmTime.toFixed(2)}s)</span>
                    </div>
                    <div className="history-output">{entry.llmOutput}</div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {activeTab === 'prompt' && (
        <section className="section">
          <h2>LLM Prompt Template</h2>
          <p className="hint">
            Customize the prompt sent to Ollama. Use <code>{'{{transcript}}'}</code> as a placeholder for the raw whisper output.
          </p>
          <div className="form-group">
            <textarea
              className="prompt-editor"
              value={config.llmPrompt || defaultPrompt}
              onChange={(e) => saveConfig({ llmPrompt: e.target.value })}
              rows={8}
            />
          </div>
          <button 
            className="btn-secondary"
            onClick={() => saveConfig({ llmPrompt: defaultPrompt })}
          >
            Reset to Default
          </button>
        </section>
      )}
    </div>
  );
}

export default App;
