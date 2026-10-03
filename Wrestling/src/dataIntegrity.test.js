import { categories, allTechniques, levels } from './data/wrestling';
import { equipment, readiness, trainingStages, homeSessions } from './data/training';

test('Wrestling catalog keeps the expected product shape', () => {
  expect(categories).toHaveLength(3);
  expect(levels).toEqual(expect.arrayContaining(['Все', 'Лёгкий', 'Средний', 'Продвинутый']));
  expect(allTechniques.length).toBeGreaterThanOrEqual(250);
  expect(new Set(allTechniques.map(x => x.name)).size).toBe(allTechniques.length);
});

test('home training data has staged sessions and required equipment', () => {
  expect(trainingStages.length).toBe(3);
  expect(homeSessions.length).toBeGreaterThanOrEqual(3);
  expect(equipment.length).toBeGreaterThan(0);
  expect(readiness.length).toBeGreaterThan(0);
});
