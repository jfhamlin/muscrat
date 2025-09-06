export class SymbolsCache {
  constructor(getSymbolsFunc) {
    this.getSymbolsFunc = getSymbolsFunc;
    this.symbols = [];
    this.symbolsByName = new Map();
    this.symbolsByGroup = new Map();
    this.initialized = false;
  }

  async initialize() {
    if (this.initialized) {
      return;
    }

    try {
      const rawSymbols = await this.getSymbolsFunc();
      this.processSymbols(rawSymbols);
      this.initialized = true;
    } catch (error) {
      console.error('Failed to fetch Muscrat symbols:', error);
      throw error;
    }
  }

  processSymbols(rawSymbols) {
    this.symbols = rawSymbols;
    
    rawSymbols.forEach(symbol => {
      this.symbolsByName.set(symbol.name, symbol);
      
      const group = symbol.group || 'Uncategorized';
      if (!this.symbolsByGroup.has(group)) {
        this.symbolsByGroup.set(group, []);
      }
      this.symbolsByGroup.get(group).push(symbol);
    });
  }

  getAllSymbols() {
    return this.symbols;
  }

  getSymbol(name) {
    return this.symbolsByName.get(name);
  }

  getSymbolsByGroup(group) {
    return this.symbolsByGroup.get(group) || [];
  }

  searchSymbols(prefix) {
    if (!prefix) {
      return this.symbols;
    }
    
    const lowerPrefix = prefix.toLowerCase();
    return this.symbols.filter(symbol => 
      symbol.name.toLowerCase().startsWith(lowerPrefix)
    );
  }

  getGroups() {
    return Array.from(this.symbolsByGroup.keys());
  }
}