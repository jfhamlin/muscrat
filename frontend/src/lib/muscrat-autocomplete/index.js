import { createCompletionProvider, createHoverProvider } from './completionProvider';
import { SymbolsCache } from './symbolsCache';

let isInitialized = false;
let completionProviderDisposable = null;
let hoverProviderDisposable = null;

export async function setupMuscratAutocomplete(monaco, getSymbolsFunc) {
  if (isInitialized) {
    console.warn('Muscrat autocomplete already initialized');
    return;
  }

  try {
    // Configure language to recognize hyphens as part of words
    monaco.languages.setLanguageConfiguration('clojure', {
      wordPattern: /[a-zA-Z0-9_\-\*\+\!\?\<\>\=]+/,
      brackets: [
        ['(', ')'],
        ['[', ']'],
        ['{', '}']
      ],
      autoClosingPairs: [
        { open: '(', close: ')' },
        { open: '[', close: ']' },
        { open: '{', close: '}' },
        { open: '"', close: '"' }
      ],
      surroundingPairs: [
        { open: '(', close: ')' },
        { open: '[', close: ']' },
        { open: '{', close: '}' },
        { open: '"', close: '"' }
      ]
    });
    
    const symbolsCache = new SymbolsCache(getSymbolsFunc);
    await symbolsCache.initialize();
    
    // Register completion provider
    const completionProvider = createCompletionProvider(monaco, symbolsCache);
    completionProviderDisposable = monaco.languages.registerCompletionItemProvider(
      'clojure',
      completionProvider
    );
    
    // Register hover provider
    const hoverProvider = createHoverProvider(monaco, symbolsCache);
    hoverProviderDisposable = monaco.languages.registerHoverProvider(
      'clojure',
      hoverProvider
    );
    
    isInitialized = true;
    console.log('Muscrat autocomplete and hover providers initialized successfully');
  } catch (error) {
    console.error('Failed to initialize Muscrat autocomplete:', error);
  }
}

export function disposeMuscratAutocomplete() {
  if (completionProviderDisposable) {
    completionProviderDisposable.dispose();
    completionProviderDisposable = null;
  }
  if (hoverProviderDisposable) {
    hoverProviderDisposable.dispose();
    hoverProviderDisposable = null;
  }
  isInitialized = false;
}