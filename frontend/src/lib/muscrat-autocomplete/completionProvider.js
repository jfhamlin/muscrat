import {
  formatDocumentation,
  formatDetail,
  generateInsertText,
  getSortText,
  shouldTriggerCompletion,
  extractPrefix,
  getCurrentFunctionContext
} from './completionUtils';

export function createCompletionProvider(monaco, symbolsCache) {
  return {
    triggerCharacters: ['(', ':', ' '],
    
    provideCompletionItems: async (model, position, context, token) => {
      try {
        const lineContent = model.getLineContent(position.lineNumber);
        const textBeforeCursor = lineContent.substring(0, position.column - 1);
        
        if (!shouldTriggerCompletion(context, textBeforeCursor)) {
          return { suggestions: [] };
        }
        
        const prefixInfo = extractPrefix(textBeforeCursor);
        let suggestions = [];
        
        const wordRange = {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: position.column - prefixInfo.prefix.length,
          endColumn: position.column
        };
        
        if (prefixInfo.type === 'keyword') {
          // Provide keyword parameter suggestions
          const currentFunction = getCurrentFunctionContext(textBeforeCursor);
          if (currentFunction) {
            suggestions = getParameterSuggestions(monaco, symbolsCache, currentFunction, prefixInfo.prefix, position, model);
          }
        } else if (prefixInfo.type === 'parameters-only') {
          // Only show parameters for the current function
          const currentFunction = getCurrentFunctionContext(textBeforeCursor);
          if (currentFunction) {
            suggestions = getParameterSuggestions(monaco, symbolsCache, currentFunction, prefixInfo.prefix, position, model);
          }
        } else if (prefixInfo.type === 'mixed') {
          // Provide both function and parameter suggestions
          const currentFunction = getCurrentFunctionContext(textBeforeCursor);
          
          // Get function suggestions
          const functionSuggestions = getFunctionSuggestions(monaco, symbolsCache, prefixInfo.prefix, wordRange);
          
          // Get parameter suggestions if we're in a function context
          let paramSuggestions = [];
          if (currentFunction) {
            paramSuggestions = getParameterSuggestions(monaco, symbolsCache, currentFunction, prefixInfo.prefix, position, model);
          }
          
          // Combine suggestions, parameters first
          suggestions = [...paramSuggestions, ...functionSuggestions];
        } else {
          // Provide function suggestions only
          suggestions = getFunctionSuggestions(monaco, symbolsCache, prefixInfo.prefix, wordRange);
        }
        
        return { suggestions };
      } catch (error) {
        console.error('Error providing completions:', error);
        return { suggestions: [] };
      }
    }
  };
}

function getFunctionSuggestions(monaco, symbolsCache, prefix, wordRange) {
  const matchingSymbols = symbolsCache.searchSymbols(prefix);
  
  return matchingSymbols.map(symbol => {
    const label = symbol.name;
    const labelDetail = formatLabelDetail(symbol);
    
    return {
      label: {
        label: label,
        detail: labelDetail,
        description: symbol.doc ? symbol.doc.substring(0, 80) + (symbol.doc.length > 80 ? '...' : '') : undefined
      },
      kind: getCompletionItemKind(monaco, symbol),
      insertText: generateInsertText(symbol),
      insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
      documentation: {
        value: formatDocumentation(symbol),
        isTrusted: true
      },
      detail: formatDetail(symbol),
      sortText: getSortText(symbol),
      range: wordRange,
      preselect: false
    };
  });
}

function formatLabelDetail(symbol) {
  if (symbol.ugenargs && symbol.ugenargs.length > 0) {
    const params = symbol.ugenargs.slice(0, 3).map(arg => arg.name).join(' ');
    const more = symbol.ugenargs.length > 3 ? ' ...' : '';
    return ` (${params}${more})`;
  }
  
  if (symbol.arglists && symbol.arglists.length > 0) {
    const firstArglist = symbol.arglists[0];
    if (Array.isArray(firstArglist) && firstArglist.length > 1) {
      const args = firstArglist.slice(1, 4).join(' ');
      const more = firstArglist.length > 4 ? ' ...' : '';
      return ` (${args}${more})`;
    }
  }
  
  return '';
}

function getParameterSuggestions(monaco, symbolsCache, functionName, prefix, position, model) {
  const symbol = symbolsCache.getSymbol(functionName);
  if (!symbol || !symbol.ugenargs) {
    return [];
  }
  
  const wordRange = {
    startLineNumber: position.lineNumber,
    endLineNumber: position.lineNumber,
    startColumn: position.column - prefix.length,
    endColumn: position.column
  };
  
  // Use a Set to track unique parameter names
  const seenParams = new Set();
  const uniqueParams = [];
  
  for (const arg of symbol.ugenargs) {
    if (!seenParams.has(arg.name)) {
      seenParams.add(arg.name);
      uniqueParams.push(arg);
    }
  }
  
  return uniqueParams
    .filter(arg => arg.name.toLowerCase().startsWith(prefix.toLowerCase()))
    .map(arg => {
      const defaultText = arg.default !== null && arg.default !== undefined 
        ? ` (default: ${arg.default})` 
        : '';
      
      // Check if we already have a colon in the prefix (from typing :param)
      const lineContent = model.getLineContent(position.lineNumber);
      const hasColon = position.column > 1 && 
        lineContent.charAt(position.column - prefix.length - 2) === ':';
      
      return {
        label: {
          label: `:${arg.name}`,
          detail: defaultText,
          description: arg.doc || ''
        },
        kind: monaco.languages.CompletionItemKind.Property,
        insertText: hasColon ? `${arg.name} \${1:value}` : `:${arg.name} \${1:value}`,
        insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
        documentation: {
          value: `**${arg.name}**${defaultText}\n\n${arg.doc || 'No description available'}`,
          isTrusted: true
        },
        sortText: arg.default !== null && arg.default !== undefined ? '1' + arg.name : '0' + arg.name,
        range: wordRange
      };
    });
}

function getCompletionItemKind(monaco, symbol) {
  if (!monaco.languages.CompletionItemKind) {
    return undefined;
  }
  
  if (symbol.name.startsWith('def')) {
    return monaco.languages.CompletionItemKind.Keyword;
  }
  
  if (symbol.group === 'Constants') {
    return monaco.languages.CompletionItemKind.Constant;
  }
  
  if (symbol.arglists || symbol.ugenargs) {
    return monaco.languages.CompletionItemKind.Function;
  }
  
  return monaco.languages.CompletionItemKind.Variable;
}

export function createHoverProvider(monaco, symbolsCache) {
  return {
    provideHover: (model, position) => {
      try {
        // Get word at position, including hyphens
        let word = model.getWordAtPosition(position);
        
        // If no word found or it doesn't include hyphens, try custom extraction
        if (!word || !word.word.includes('-')) {
          const lineContent = model.getLineContent(position.lineNumber);
          const match = lineContent.substring(0, position.column).match(/([a-z-]+)$/i);
          if (match) {
            const startCol = position.column - match[1].length;
            const endMatch = lineContent.substring(position.column - 1).match(/^([a-z-]*)/i);
            const fullWord = match[1] + (endMatch ? endMatch[1] : '');
            word = {
              word: fullWord,
              startColumn: startCol,
              endColumn: startCol + fullWord.length
            };
          }
        }
        
        if (!word) {
          return null;
        }
        
        const symbol = symbolsCache.getSymbol(word.word);
        if (!symbol) {
          return null;
        }
        
        const documentation = formatDocumentation(symbol);
        const signature = formatDetail(symbol);
        
        return {
          contents: [
            { value: `**${symbol.name}**`, isTrusted: true },
            { value: signature, isTrusted: true },
            { value: documentation, isTrusted: true }
          ],
          range: new monaco.Range(
            position.lineNumber,
            word.startColumn,
            position.lineNumber,
            word.endColumn
          )
        };
      } catch (error) {
        console.error('Error providing hover:', error);
        return null;
      }
    }
  };
}