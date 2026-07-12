import { describe, expect, it } from 'vitest';
import router from './index';

describe('router', () => {
  it('registers the create resource route', () => {
    const route = router.getRoutes().find((item) => item.name === 'CreateResource');

    expect(route).toBeTruthy();
    expect(route?.path).toBe('/resources/create');
  });
});
