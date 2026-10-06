import React, { useMemo, useState } from 'react';
import { categories, levels, allTechniques } from './data/wrestling';
import { trainingStages, progressionRules, progressSteps, homeSessions, equipment, readiness, weeklyPlan } from './data/training';
import './styles.css';

const navItems = [...categories.map(c => c.id), 'тренировки', 'прогресс'];
const youtube = q => `https://www.youtube.com/results?search_query=${encodeURIComponent(q + ' wrestling tutorial')}`;
const imageSearch = q => `https://www.google.com/search?tbm=isch&q=${encodeURIComponent(q + ' wrestling technique')}`;

function load(key){ try { return JSON.parse(localStorage.getItem(key) || '[]'); } catch { return []; } }

export default function App(){
  const [page,setPage] = useState('главная');
  const [filters,setFilters] = useState({group:'Все',level:'Все'});
  const [query,setQuery] = useState('');
  const [selected,setSelected] = useState(null);
  const [theme,setTheme] = useState(() => localStorage.getItem('wrestling-theme') || 'night');
  const [favorites,setFavorites] = useState(() => load('wrestling-favorites'));
  const [completed,setCompleted] = useState(() => load('wrestling-completed'));
  const [techCompleted,setTechCompleted] = useState(() => load('wrestling-tech-completed'));
  const [equipmentDone,setEquipmentDone] = useState(() => load('wrestling-equipment'));
  const [readinessDone,setReadinessDone] = useState(() => load('wrestling-readiness'));
  const [menuOpen,setMenuOpen] = useState(false);

  const toggle=(key,setter,storage)=>setter(prev=>{const next=prev.includes(key)?prev.filter(x=>x!==key):[...prev,key];localStorage.setItem(storage,JSON.stringify(next));return next;});
  const active=categories.find(c=>c.id===page);
  const items=useMemo(()=>{
    if(!active)return[];

    const normalize = value => String(value ?? '')
      .toLowerCase()
      .replace(/ё/g,'е')
      .normalize('NFD')
      .replace(/[\\u0300-\\u036f]/g,'')
      .replace(/[^a-zа-я0-9]+/gi,' ')
      .trim();

    const distance = (a,b) => {
      if(a===b)return 0;
      if(!a)return b.length;
      if(!b)return a.length;
      const prev=Array.from({length:b.length+1},(_,i)=>i);
      for(let i=1;i<=a.length;i++){
        let left=i;
        for(let j=1;j<=b.length;j++){
          const up=prev[j];
          const cost=a[i-1]===b[j-1]?0:1;
          prev[j]=Math.min(prev[j]+1,left+1,prev[j-1]+cost);
          left=up;
        }
      }
      return prev[b.length];
    };

    const searchable = active.groups.flatMap(g=>g.items.map(([name,level,english,text,lesson])=>({
      name,level,english,text,lesson,group:g.title,jp:g.jp,
      video:youtube(english),image:imageSearch(english)
    })));

    const filtered=searchable.filter(x=>
      (filters.group==='Все'||x.group===filters.group) &&
      (filters.level==='Все'||x.level===filters.level)
    );

    const q=normalize(query);
    if(!q)return filtered;

    const words=q.split(/\\s+/).filter(Boolean);

    return filtered
      .map(item=>{
        const fields=[
          [normalize(item.name),120],
          [normalize(item.english),95],
          [normalize(item.group),70],
          [normalize(item.level),55],
          [normalize(item.text),45],
          [normalize(item.lesson),35],
          [normalize(item.jp),25]
        ];
        const full=fields.map(([v])=>v).join(' ');
        let score=0;

        if(full===q)score+=300;
        if(full.includes(q))score+=180;

        for(const word of words){
          let best=0;
          for(const [value,weight] of fields){
            if(!value)continue;
            if(value===word)best=Math.max(best,weight+70);
            else if(value.split(' ').some(token=>token.startsWith(word)))best=Math.max(best,weight+45);
            else if(value.includes(word))best=Math.max(best,weight);
            else {
              const tokens=value.split(' ').filter(Boolean);
              const near=tokens.some(token=>{
                const maxDistance=word.length<=4?1:word.length<=7?2:3;
                return Math.abs(token.length-word.length)<=maxDistance && distance(token,word)<=maxDistance;
              });
              if(near)best=Math.max(best,weight*0.55);
            }
          }
          score+=best;
        }

        return {...item,score};
      })
      .filter(item=>item.score>0)
      .sort((a,b)=>b.score-a.score);
  },[active,filters,query]);

  const go=next=>{setPage(next);setQuery('');setFilters({group:'Все',level:'Все'});setMenuOpen(false);window.scrollTo({top:0,behavior:'smooth'});};
  const changeTheme=()=>{const next=theme==='night'?'day':'night';setTheme(next);localStorage.setItem('wrestling-theme',next);};

  return <div className={`site ${theme}`}>
    <aside className="sidebar">
      <button className="brand" onClick={()=>go('главная')}><span>闘</span><b>W</b><small>WRESTLING</small></button>
      <nav className="sideNav">
        <button className={page==='главная'?'active':''} onClick={()=>go('главная')}><i>⌂</i><span>Главная</span></button>
        {navItems.map(item=><button key={item} className={page===item?'active':''} onClick={()=>go(item)}><i>{item==='тренировки'?'◈':item==='прогресс'?'◒':'武'}</i><span>{item==='тренировки'?'Тренировки':item==='прогресс'?'Прогресс':categories.find(c=>c.id===item)?.title}</span></button>)}
      </nav>
      <div className="sideBottom">
        <button onClick={changeTheme}>{theme==='night'?'☼':'☾'} <span>{theme==='night'?'День':'Ночь'}</span></button>
        <div>★ {favorites.length}</div>
      </div>
    </aside>
    <header className="mobileHeader">
      <button className="mobileBrand" onClick={()=>go('главная')}><span>闘</span> WRESTLING</button>
      <button onClick={changeTheme}>{theme==='night'?'☼':'☾'}</button>
      <button onClick={()=>setMenuOpen(v=>!v)}>☰</button>
    </header>
    {menuOpen&&<div className="mobileMenu">{['главная',...navItems].map(item=><button key={item} className={page===item?'active':''} onClick={()=>go(item)}>{item==='главная'?'Главная':item==='тренировки'?'Тренировки':item==='прогресс'?'Прогресс':categories.find(c=>c.id===item)?.title}</button>)}</div>}

    {page==='главная'?<Home go={go}/>:page==='тренировки'?<Training completed={completed} toggleCompleted={key=>toggle(key,setCompleted,'wrestling-completed')} equipmentDone={equipmentDone} toggleEquipment={key=>toggle(key,setEquipmentDone,'wrestling-equipment')} readinessDone={readinessDone} toggleReadiness={key=>toggle(key,setReadinessDone,'wrestling-readiness')}/>:page==='прогресс'?<Progress completed={completed} techCompleted={techCompleted} total={allTechniques.length} equipmentDone={equipmentDone} readinessDone={readinessDone}/>:<Catalog active={active} filters={filters} setFilters={setFilters} query={query} setQuery={setQuery} items={items} favorites={favorites} toggleFavorite={key=>toggle(key,setFavorites,'wrestling-favorites')} setSelected={setSelected}/>}
    {selected&&<Modal item={selected} favorite={favorites.includes(selected.name)} techniqueDone={techCompleted.includes(selected.name)} toggleTechnique={()=>toggle(selected.name,setTechCompleted,'wrestling-tech-completed')} toggleFavorite={()=>toggle(selected.name,setFavorites,'wrestling-favorites')} onClose={()=>setSelected(null)}/>}
  </div>;
}

function Home({go}){
  return <main className="content">
    <section className="heroExact">
      <div className="heroOverlay"/>
      <div className="heroText">
        <span className="kicker">01 / ОБЗОР</span>
        <h1>Wrestling</h1>
        <p>Техника, партер, ОФП, тренировки, прогресс и адаптация под экран.</p>
        <div className="heroButtons"><button className="redBtn" onClick={()=>go('стойка')}>Открыть технику</button><button className="lightBtn" onClick={()=>go('тренировки')}>Домашняя тренировка</button></div>
      </div>
      <div className="heroVertical">柔道 · レスリング</div>
    </section>
    <section className="threeCols">
      {categories.map((c,i)=><button key={c.id} className="sectionCard" onClick={()=>go(c.id)}>
        <div className="sectionCardTop"><span>0{i+1}</span><b>{c.jp}</b></div><h2>{c.title}</h2><p>{c.description}</p><strong>{c.groups.reduce((n,g)=>n+g.items.length,0)} элементов</strong>
      </button>)}
    </section>
    <section className="overviewGrid">
      <article className="featureCard large"><span className="kicker">01 / СТАРТ</span><h2>Универсальная стойка</h2><p>Положение, перемещения, изменение уровня, проходы, броски, подсечки, подножки и зацепы.</p><button onClick={()=>go('стойка')}>Перейти →</button></article>
      <article className="featureCard"><span className="kicker">02 / ПАРТЕР</span><h2>Работа на земле</h2><p>Позиции, перевороты, удержания, удушающие и болевые с контролем безопасности.</p><button onClick={()=>go('партер')}>Открыть →</button></article>
      <article className="featureCard"><span className="kicker">03 / ОФП</span><h2>Подготовка</h2><p>База силы, выносливости, мобильности и работа с доступным оборудованием.</p><button onClick={()=>go('ОФП')}>Открыть →</button></article>
    </section>
    <section className="bottomGrid">
      <article><span className="kicker">ДОМАШНЯЯ ТРЕНИРОВКА</span><h2>Три этапа дома</h2><p>База → связки → раунды. С понятным оборудованием и отметками прогресса.</p><button onClick={()=>go('тренировки')}>Начать →</button></article>
      <article><span className="kicker">ПРОГРЕСС</span><h2>Смотри динамику</h2><p>Техника, тренировки, готовность и оборудование — в одном месте.</p><button onClick={()=>go('прогресс')}>Открыть →</button></article>
      <article><span className="kicker">ЭКИПИРОВКА</span><h2>Минимальный набор</h2><p>Безопасное покрытие, резина, вода и то, что реально пригодится дома.</p><span className="verticalMark">技 · дисциплина · уважение</span></article>
    </section>
  </main>;
}

function Catalog({active,filters,setFilters,query,setQuery,items,favorites,toggleFavorite,setSelected}){
  return <main className="content">
    <section className="pageTitle"><span className="kicker">WRESTLING / {active.jp}</span><h1>{active.title}</h1><p>{active.description}</p></section>
    <section className="filters"><div className="search"><span>⌕</span><input value={query} onChange={e=>setQuery(e.target.value)} placeholder="Поиск техники…"/></div><div className="chips"><button className={filters.group==='Все'?'on':''} onClick={()=>setFilters({...filters,group:'Все'})}>Все</button>{active.groups.map(g=><button key={g.id} className={filters.group===g.title?'on':''} onClick={()=>setFilters({...filters,group:g.title})}>{g.title}</button>)}</div><div className="chips">{levels.map(l=><button key={l} className={filters.level===l?'on':''} onClick={()=>setFilters({...filters,level:l})}>{l}</button>)}</div></section>
    {query && <div className="searchResultLine">Найдено: <b>{items.length}</b> · запрос: <span>{query}</span></div>}
    <section className="catalogCards">{items.map((item,i)=><article className="techCard" key={item.name}><div className="techImage"><span>{String(i+1).padStart(2,'0')}</span><b>{item.jp}</b><strong>闘</strong><em>{item.level}</em></div><div className="techBody"><small>{item.group} · {item.english}</small><h2>{item.name}</h2><p>{item.text}</p><div><button onClick={()=>setSelected(item)}>Разбор</button><a href={item.video} target="_blank" rel="noreferrer">Видео</a><button className="fav" onClick={()=>toggleFavorite(item.name)}>{favorites.includes(item.name)?'★':'☆'}</button></div></div></article>)}</section>
    {query && !items.length && <section className="emptySearch"><b>Ничего не найдено</b><p>Попробуй более короткий запрос, другое написание или выбери «Все» в фильтрах.</p></section>}
  </main>;
}

function Modal({item,favorite,techniqueDone,toggleTechnique,toggleFavorite,onClose}){
  return <div className="modalShade" onClick={onClose}><div className="modalExact" onClick={e=>e.stopPropagation()}><button className="modalClose" onClick={onClose}>×</button><div className="modalImage"><span>{item.jp}</span><strong>闘</strong></div><div className="modalText"><span className="kicker">{item.group} · {item.english}</span><h2>{item.name}</h2><p>{item.text}</p><div className="lessonBox"><b>ОБУЧЕНИЕ</b><p>{item.lesson}</p><ol><li>Сначала положение и движение без сопротивления.</li><li>Затем медленные качественные повторы.</li><li>После стабильной техники — лёгкое сопротивление.</li><li>Опасные элементы только с тренером и подготовленным партнёром.</li></ol></div><div className="modalActions"><a className="redBtn" href={item.video} target="_blank" rel="noreferrer">Смотреть обучение →</a><button onClick={toggleTechnique}>{techniqueDone?'✓ Изучено':'○ Отметить изученным'}</button><button onClick={toggleFavorite}>{favorite?'★ В избранном':'☆ В избранное'}</button></div></div></div></div>;
}

function Training({completed,toggleCompleted,equipmentDone,toggleEquipment,readinessDone,toggleReadiness}){
  return <main className="content"><section className="pageTitle"><span className="kicker">WRESTLING / ДОМА</span><h1>Тренировки</h1><p>Три этапа дома: база, связки и раунды. Оборудование — только то, что реально использовать дома.</p></section><section className="trainingGrid">{trainingStages.map(stage=><article className="trainingCard" key={stage.id}><span className="step">0{stage.id}</span><h2>{stage.title}</h2><b>{stage.duration}</b><p>{stage.goal}</p><ul>{stage.blocks.map((block,i)=><li key={i}><button className={completed.includes(stage.id+':'+i)?'done':''} onClick={()=>toggleCompleted(stage.id+':'+i)}>✓</button>{block}</li>)}</ul></article>)}</section><section className="homeRows"><article><span className="kicker">ГОТОВНОСТЬ</span><h2>Перед тренировкой</h2><div className="checkGrid">{readiness.map((item,i)=><button key={i} className={readinessDone.includes(String(i))?'done':''} onClick={()=>toggleReadiness(String(i))}>{readinessDone.includes(String(i))?'✓':'○'} {item}</button>)}</div></article><article><span className="kicker">ЭКИПИРОВКА</span><h2>Что нужно дома</h2><div className="checkGrid">{equipment.map(item=><button key={item.id} className={equipmentDone.includes(item.id)?'done':''} onClick={()=>toggleEquipment(item.id)}>{equipmentDone.includes(item.id)?'✓':'○'} {item.name}</button>)}</div></article></section></main>;
}

function Progress({completed,techCompleted,total,equipmentDone,readinessDone}){
  const workoutTotal=trainingStages.reduce((n,s)=>n+s.blocks.length,0);
  const vals=[Math.min(100,Math.round(completed.length/Math.max(workoutTotal,1)*100)),Math.min(100,Math.round(techCompleted.length/Math.max(total,1)*100)),Math.round(readinessDone.length/readiness.length*100),Math.round(equipmentDone.length/equipment.length*100)];
  return <main className="content"><section className="pageTitle"><span className="kicker">WRESTLING / ПРОГРЕСС</span><h1>Прогресс</h1><p>Следи за тренировками, изученной техникой, готовностью и регулярностью.</p></section><section className="progressGrid">{['Тренировки','Техника','Готовность','Оборудование'].map((x,i)=><article key={x}><strong>{vals[i]}%</strong><span>{x}</span><div><i style={{width:`${vals[i]}%`}}/></div></article>)}</section><section className="progressWide"><h2>Последовательность важнее скорости.</h2><p>Сначала качество движения и безопасность, затем объём и сопротивление.</p></section><section className="steps">{progressSteps.map((s,i)=><article key={s.title}><span>0{i+1}</span><div><h2>{s.title}</h2><p>{s.text}</p></div></article>)}</section></main>;
}
