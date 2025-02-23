import { useState } from 'react';


export default ({ options, selected, onSelect }) => {

  return (
    <div className="flex flex-row gap-10 text-gray-600 border-b border-gray-600 mb-2">
      {options.map((option, index) => (
        <Button key={index} onClick={() => onSelect(option)} selected={selected === option}>{option.toUpperCase()}</Button>
      ))}
    </div>
  );
};

const Button = ({ children, selected, onClick }) => {
  const colorClass = selected ? ' text-gray-200 border-gray-200 border-b box-content' : '';
  return (
    <div className={'text-xs cursor-pointer' + colorClass} onClick={onClick}>
      {children}
    </div>
  );
}
