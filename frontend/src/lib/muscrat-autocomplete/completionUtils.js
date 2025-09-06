const GROUP_PRIORITY = {
  'Core': '1',
  'Oscillators': '2',
  'Effects': '3',
  'Filters': '4',
  'Envelopes': '5',
  'Math': '6',
  'Utility': '7',
  'Constants': '8',
  'Uncategorized': '9'
};

export function formatArglist(arglist) {
  if (!arglist) return '';
  
  if (Array.isArray(arglist) && arglist.length > 0) {
    const args = arglist[0];
    if (typeof args === 'string') {
      return args;
    }
    if (Array.isArray(args)) {
      return args.join(' ');
    }
  }
  return '';
}

export function generateInsertText(symbol) {
  const name = symbol.name;
  
  if (symbol.ugenargs && symbol.ugenargs.length > 0) {
    const requiredArgs = [];
    const optionalArgs = [];
    
    symbol.ugenargs.forEach((arg, index) => {
      if (arg.name === 'in' || arg.name === 'freq' || arg.name === 'rate' || 
          arg.name === 'signal' || arg.name === 'input') {
        requiredArgs.push(`\${${index + 1}:${arg.name}}`);
      } else if (arg.default === null || arg.default === undefined) {
        requiredArgs.push(`\${${index + 1}:${arg.name}}`);
      }
    });
    
    if (requiredArgs.length > 0) {
      return `${name} ${requiredArgs.join(' ')}`;
    }
  }
  
  if (symbol.arglists && symbol.arglists.length > 0) {
    const firstArglist = symbol.arglists[0];
    if (Array.isArray(firstArglist) && firstArglist.length > 0) {
      const args = firstArglist.slice(1);
      if (args.length > 0) {
        const snippetArgs = args.map((arg, i) => 
          `\${${i + 1}:${typeof arg === 'string' ? arg : 'arg' + (i + 1)}}`
        );
        return `${name} ${snippetArgs.join(' ')}`;
      }
    }
  }
  
  return name;
}

export function formatDocumentation(symbol) {
  let doc = '';
  
  if (symbol.doc) {
    doc = symbol.doc;
  }
  
  if (symbol.ugenargs && symbol.ugenargs.length > 0) {
    doc += '\n\n**Parameters:**\n';
    symbol.ugenargs.forEach(arg => {
      const defaultVal = arg.default !== null && arg.default !== undefined 
        ? ` (default: ${arg.default})` 
        : '';
      doc += `- **${arg.name}**${defaultVal}: ${arg.doc || 'No description'}\n`;
    });
  }
  
  return doc;
}

export function formatDetail(symbol) {
  if (symbol.arglists && symbol.arglists.length > 0) {
    const signatures = symbol.arglists.map(arglist => {
      if (Array.isArray(arglist)) {
        return `(${symbol.name} ${arglist.slice(1).join(' ')})`;
      }
      return `(${symbol.name})`;
    });
    return signatures.join('\n');
  }
  
  if (symbol.ugenargs && symbol.ugenargs.length > 0) {
    const args = symbol.ugenargs.map(arg => arg.name).join(' ');
    return `(${symbol.name} ${args})`;
  }
  
  return `(${symbol.name})`;
}

export function getSortText(symbol) {
  const groupPriority = GROUP_PRIORITY[symbol.group] || '9';
  return groupPriority + symbol.name;
}

export function shouldTriggerCompletion(context, textBeforeCursor) {
  // Always trigger on explicit request
  if (context.triggerKind === 1) {
    return true;
  }
  
  // Trigger on specific characters
  if (context.triggerCharacter === '(' || context.triggerCharacter === ':' || context.triggerCharacter === ' ') {
    return true;
  }
  
  // Check if we're inside unclosed parentheses
  if (isInsideParentheses(textBeforeCursor)) {
    return true;
  }
  
  return false;
}

export function isInsideParentheses(text) {
  let depth = 0;
  for (let i = 0; i < text.length; i++) {
    if (text[i] === '(') depth++;
    else if (text[i] === ')') depth--;
  }
  return depth > 0;
}

export function extractPrefix(textBeforeCursor) {
  // Check for keyword parameter
  const keywordMatch = textBeforeCursor.match(/:([a-z-]*)$/i);
  if (keywordMatch) {
    return { type: 'keyword', prefix: keywordMatch[1] };
  }
  
  // Check for function after opening paren
  const match = textBeforeCursor.match(/\(([a-z-]*)$/i);
  if (match) {
    return { type: 'function', prefix: match[1] };
  }
  
  // Check for word at cursor (could be function or value)
  const wordMatch = textBeforeCursor.match(/([a-z-]+)$/i);
  if (wordMatch) {
    // Determine context to decide if it's a function or parameters
    const context = getCompletionContext(textBeforeCursor);
    return { type: context, prefix: wordMatch[1] };
  }
  
  // Default based on context when we're inside parens
  if (isInsideParentheses(textBeforeCursor)) {
    const context = getCompletionContext(textBeforeCursor);
    return { type: context, prefix: '' };
  }
  
  return { type: 'function', prefix: '' };
}

export function getCompletionContext(textBeforeCursor) {
  // Remove the current word being typed to analyze context
  const withoutCurrentWord = textBeforeCursor.replace(/[a-z-]+$/i, '');
  
  // Check if we're directly after a function name (e.g., "(sin " or "(sin 440 ")
  const currentFunc = getCurrentFunctionContext(textBeforeCursor);
  if (currentFunc) {
    // We're inside a function call - only show parameters for that function
    // Unless we're starting a new nested expression with (
    if (withoutCurrentWord.match(/\($/) || withoutCurrentWord.match(/\(\s*$/)) {
      return 'function'; // Starting a new nested function
    }
    return 'parameters-only'; // Show only parameters for current function
  }
  
  // If we're right after opening paren, suggest functions
  if (withoutCurrentWord.match(/\($/) || withoutCurrentWord.match(/\(\s+$/)) {
    return 'function';
  }
  
  return 'function';
}

export function getCurrentFunctionContext(textBeforeCursor) {
  // Find the most recent unclosed opening parenthesis and extract function name
  let depth = 0;
  let lastOpenIndex = -1;
  
  for (let i = textBeforeCursor.length - 1; i >= 0; i--) {
    if (textBeforeCursor[i] === ')') {
      depth++;
    } else if (textBeforeCursor[i] === '(') {
      if (depth === 0) {
        lastOpenIndex = i;
        break;
      }
      depth--;
    }
  }
  
  if (lastOpenIndex === -1) {
    return null;
  }
  
  // Extract function name after the opening paren
  const afterParen = textBeforeCursor.substring(lastOpenIndex + 1);
  const funcMatch = afterParen.match(/^([a-z-]+)/i);
  
  if (funcMatch) {
    return funcMatch[1];
  }
  
  return null;
}