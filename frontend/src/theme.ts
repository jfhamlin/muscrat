
interface Theme {
  background: {
    primary: string;
    secondary: string;
    tertiary: string;
  };
  accent: {
    primary: string;
  };
}

export const tailwindTheme: Theme = {
  background: {
    primary: '#1D192B',
    secondary: '#26203D',
    tertiary: '#483D6B',
  },
  accent: {
    primary: '#FFC36A',
  },
};
