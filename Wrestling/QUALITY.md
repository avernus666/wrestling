# Качество

Основные требования к Wrestling:

- понятная навигация;
- единый стиль названий и уровней;
- отсутствие дублирования учебных материалов;
- корректная работа фильтров;
- сохранение избранного и прогресса;
- адаптация интерфейса под мобильные экраны;
- обязательное наличие предупреждений для потенциально травмоопасных техник;
- проверка ссылок на обучающие материалы перед публикацией.

## Quality baseline

Wrestling keeps the existing visual system unchanged while adding an engineering quality layer:

- React unit/integration coverage with Testing Library;
- Playwright smoke coverage for desktop and mobile;
- Go HTTP router and security middleware tests;
- catalog and training-data integrity checks;
- PWA manifest and offline shell support;
- backend health/readiness endpoints;
- API and PostgreSQL separation through services and repositories.

The visual layer is intentionally not part of this hardening pass.
