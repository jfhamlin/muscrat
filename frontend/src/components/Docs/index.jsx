import {
  useState,
  useEffect,
  useMemo,
} from 'react';

import Markdown from 'react-markdown'

import {
  GetNSPublics,
} from "../../../bindings/github.com/jfhamlin/muscrat/muscratservice";

import 'github-markdown-css';

// A public looks like this:
/* {
 *   name     string
 *   group    string
 *   doc      string
 *   arglists []any
 *   ugenargs []{ name string, doc string, default any }
 * } */

const Group = ({ group, symbols, selectItem, selectedItemName }) => {
  const [collapsed, setCollapsed] = useState(true);

  return (
    <div>
      {/* gray background, white text, padding, margin, cursor pointer */}
      <h2 className="bg-gray-500 text-white p-2 my-1 cursor-pointer"
          onClick={() => setCollapsed(!collapsed)}>
        {group}
      </h2>
      {!collapsed && (
        <ul>
          {symbols.map((sym) => {
            return <li className={"cursor-pointer" + (selectedItemName === sym.name ? " bg-gray-300" : "")}
                       key={sym.name}
                       onClick={() => selectItem(sym)}>
              {sym.name}
            </li>
          })}
        </ul>
      )}
    </div>
  );
};

const Detail = ({ symbol }) => {
  // show the group name at the top, in small text
  // show the sym name in a large font
  // show the sym doc string
  // if ugenargs is present, show a table of ugenargs
  // else if arglists is present, show a table of arglists
  // else show nothing

  return (
    <div>
      <h3>&gt; {symbol.group}</h3>
      <h2 className="text-2xl my-2">
        {symbol.name}
      </h2>
      <div className='markdown-body'>
        <Markdown>{symbol.doc}</Markdown>
      </div>
      {symbol.ugenargs && (
        <table className="my-2 ml-5 border-separate [border-spacing:0.75rem]">
          <thead>
            <tr>
              <th>name</th>
              <th>default</th>
              <th>doc</th>
            </tr>
          </thead>
          <tbody>
            {symbol.ugenargs.map((arg) => {
              const defaultVal = arg.default === null ? "nil" :
                                 (arg.default instanceof Object ? '?' : arg.default);
              return <tr key={arg.name}>
                <td><pre>{arg.name}</pre></td>
                <td>{defaultVal}</td>
                <td>{arg.doc}</td>
              </tr>
            })}
          </tbody>
        </table>
      )}
    </div>
  );
};

const Directory = ({ symbols, selectItem, selectedItemName }) => {
  const groups = useMemo(() => {
    return symbols.reduce((acc, curr) => {
      if (!acc[curr.group]) {
        acc[curr.group] = [];
      }
      acc[curr.group].push(curr);
      return acc;
    }, {});
  }, [symbols]);

  return (
    <div className="select-none">
      {Object.keys(groups).map((group) => {
        return (
          <Group key={group}
                 group={group}
                 symbols={groups[group]}
                 selectItem={selectItem}
                 selectedItemName={selectedItemName}
          />
        );
      })}
    </div>
  );
};

export default () => {
  const [symbols, setSymbols] = useState([]);
  const [selectedSymbol, setSelectedSymbol] = useState(null);

  useEffect(() => {
    GetNSPublics().then((res) => {
      setSymbols(res);
    });
  }, []);

  const groups = useMemo(() => {
    return symbols.reduce((acc, curr) => {
      if (!acc[curr.group]) {
        acc[curr.group] = [];
      }
      acc[curr.group].push(curr);
      return acc;
    }, {});
  }, [symbols]);

  return (
    // outer div has fixed height, inner divs scroll
    <div className="pb-4 mx-2 h-full overflow-hidden mb-2 text-gray-200">
      <div className="flex overflow-hidden h-full">
        {/* each child scrolls separately */}
        <div className="w-100 overflow-y-auto">
          <Directory symbols={symbols}
                     selectedItemName={selectedSymbol ? selectedSymbol.name : null}
                     selectItem={setSelectedSymbol} />
        </div>
        <div className="overflow-y-auto ml-2">
          {selectedSymbol && <Detail symbol={selectedSymbol} />}
        </div>
      </div>
    </div>
  );
}
