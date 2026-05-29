document.addEventListener("DOMContentLoaded", () => {
    const urlParams = new URLSearchParams(window.location.search);
    const deckId = urlParams.get('deck_id') || '1';

    const viewAllBtn = document.getElementById('viewAllBtn');
    if (viewAllBtn) {
        viewAllBtn.href = `/decks/view?deck_id=${deckId}`;
    }

    // Определяем, открыт ли сайт с мобильного устройства (ширина экрана менее 768px)
    const isMobile = window.innerWidth <= 768;

    // Адаптивные стили веера: для мобилок уменьшаем радиус разъезда карт, чтобы они влезали в экран
    const transformStyles = isMobile ? [
        'rotate(-12deg) translate(-110px, 10px)',
        'rotate(-6deg) translate(-55px, 3px)',
        'rotate(0deg) translate(0px, 0px)',
        'rotate(6deg) translate(55px, 3px)',
        'rotate(12deg) translate(110px, 10px)'
    ] : [
        'rotate(-12deg) translate(-210px, 15px)',
        'rotate(-6deg) translate(-105px, 5px)',
        'rotate(0deg) translate(0px, 0px)',
        'rotate(6deg) translate(105px, 5px)',
        'rotate(12deg) translate(210px, 15px)'
    ];

    const container = document.getElementById('cardsContainer');
    if (!container) return;

    let isIntroPlaying = true;

    transformStyles.forEach((style, idx) => {
        const card = document.createElement('div');
        card.className = `card card-${idx}`;
        card.style.transform = style;
        
        card.innerHTML = `
            <div class="card-shirt-pattern">
                FORGE
            </div>
        `;

        // Клик/Тап перенаправляет в создание карт
        card.onclick = () => {
            window.location.href = `/decks/add-cards?deck_id=${deckId}`;
        };

        // Интерактив по наведению (работает на десктопах)
        card.onmouseenter = () => {
            if (!isIntroPlaying) pushSiblings(idx, transformStyles, isMobile);
        };
        card.onmouseleave = () => {
            if (!isIntroPlaying) resetSiblings(transformStyles);
        };

        // Дополнительный интерактив для мобилок: легкое раздвижение при таче
        card.ontouchstart = () => {
            if (!isIntroPlaying) pushSiblings(idx, transformStyles, isMobile);
        };

        container.appendChild(card);
    });

    gsap.fromTo('.card', 
        { scale: 0, opacity: 0, y: 100 }, 
        { 
            scale: 1, 
            opacity: 1, 
            y: 0,
            stagger: 0.15,               
            ease: 'elastic.out(1, 0.75)', 
            delay: 0.4,
            duration: 2.0,
            onComplete: () => {
                isIntroPlaying = false;
            }
        }
    );
});

function pushSiblings(hoveredIdx, transformStyles, isMobile) {
    // На мобильных устройствах расталкиваем соседние карты слабее (на 35px вместо 70px)
    const pushDistance = isMobile ? 35 : 70;

    transformStyles.forEach((_, i) => {
        const target = document.querySelector(`.card-${i}`);
        if (!target) return;
        
        gsap.killTweensOf(target);
        const baseTransform = transformStyles[i];

        if (i === hoveredIdx) {
            const noRotation = baseTransform.replace(/rotate\([-0-9.]+deg\)/, 'rotate(0deg)').replace(/translate\(([-0-9.]+)px,\s*([-0-9.]+)px\)/, 'translate($1px, -15px) scale(1.05)');
            gsap.to(target, { transform: noRotation, duration: 0.4, ease: 'power2.out', overwrite: 'auto' });
        } else {
            const offsetX = i < hoveredIdx ? -pushDistance : pushDistance;
            const translateRegex = /translate\(([-0-9.]+)px/;
            let pushedTransform = baseTransform;
            
            if (translateRegex.test(baseTransform)) {
                const currentX = parseFloat(baseTransform.match(translateRegex)[1]);
                pushedTransform = baseTransform.replace(translateRegex, `translate(${currentX + offsetX}px`);
            }
            gsap.to(target, { transform: pushedTransform, duration: 0.4, ease: 'back.out(1.2)', overwrite: 'auto' });
        }
    });
}

function resetSiblings(transformStyles) {
    transformStyles.forEach((style, i) => {
        const target = document.querySelector(`.card-${i}`);
        if (!target) return;
        
        gsap.killTweensOf(target);
        gsap.to(target, { transform: style, duration: 0.4, ease: 'back.out(1.2)', overwrite: 'auto' });
    });
}