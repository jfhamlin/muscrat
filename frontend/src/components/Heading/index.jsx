import React from "react";

export default (props) => (
  <div className="p-1 bg-white bg-opacity-80 text-xs text-gray-500 absolute top-0 left-1/2 transform -translate-x-1/2 rounded-sm">
    {props.children}
  </div>
);
