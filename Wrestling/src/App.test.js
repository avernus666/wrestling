import React from 'react';
import { render, screen, fireEvent, within } from '@testing-library/react';
import App from './App';

beforeEach(() => {
  localStorage.clear();
  window.scrollTo = jest.fn();
});

test('renders Wrestling home and the three main sections', () => {
  render(<App />);
  expect(screen.getAllByText(/WRESTLING/i).length).toBeGreaterThan(0);
  expect(screen.getByRole('button', { name: /стойка/i })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: /партер/i })).toBeInTheDocument();
  expect(screen.getByRole('button', { name: /офп/i })).toBeInTheDocument();
});

test('opens stance catalog and filters techniques', () => {
  render(<App />);
  fireEvent.click(screen.getByRole('button', { name: /стойка/i }));
  expect(screen.getByText(/Каталог/i)).toBeInTheDocument();
  expect(screen.getByText(/элементов/i)).toBeInTheDocument();
  fireEvent.change(screen.getByPlaceholderText(/Поиск техники/i), { target: { value: 'single' } });
  expect(screen.getByText(/Low single/i)).toBeInTheDocument();
});

test('opens a technique modal and marks it as studied', () => {
  render(<App />);
  fireEvent.click(screen.getByRole('button', { name: /стойка/i }));
  const firstBreakdown = screen.getAllByRole('button', { name: /Разбор/i })[0];
  fireEvent.click(firstBreakdown);
  expect(screen.getByText(/Обучение/i)).toBeInTheDocument();
  const markButton = screen.getByRole('button', { name: /Отметить изучение|Изучено/i });
  fireEvent.click(markButton);
  expect(screen.getByRole('button', { name: /Изучено/i })).toBeInTheDocument();
});

test('theme toggle persists the selected theme', () => {
  render(<App />);
  const toggle = screen.getByRole('button', { name: /Сменить тему/i });
  fireEvent.click(toggle);
  expect(localStorage.getItem('wrestling-theme')).toBe('day');
  expect(document.querySelector('.site.day')).toBeInTheDocument();
});

test('progress page reflects locally completed training and equipment', () => {
  localStorage.setItem('wrestling-completed', JSON.stringify(['warmup-1']));
  localStorage.setItem('wrestling-equipment', JSON.stringify(['mat']));
  render(<App />);
  fireEvent.click(screen.getByRole('button', { name: /Прогресс/i }));
  expect(screen.getByText(/Домашний цикл/i)).toBeInTheDocument();
  expect(screen.getByText(/Изучение техники/i)).toBeInTheDocument();
});
