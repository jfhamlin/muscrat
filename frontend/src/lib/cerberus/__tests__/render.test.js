import Cerberus from '..';

describe('render', () => {
  it('should render a simple sphere', () => {
    const cerb = new Cerberus();

    cerb.scene().add(() => {
      cerb.sphere();
    }).render();
  });
});
