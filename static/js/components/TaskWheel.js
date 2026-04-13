import { TasksService } from '../services/TasksService.js';

export class TaskWheel {
    constructor() {
        this.tasksService = new TasksService();
        this.modalElement = null;
        this.bootstrapModal = null;
        this.tasks = [];
        this.currentRotation = 0;
        this.isSpinning = false;
        this.selectedTask = null;
        this.colors = [
            '#FF6384', // розовый
            '#36A2EB', // голубой
            '#FFCE56', // желтый
            '#4BC0C0', // бирюзовый
            '#9966FF', // фиолетовый
            '#FF9F40', // оранжевый
            '#E7E9ED', // серый
            '#8AC926', // зеленый
            '#1982C4', // синий
            '#6A4C93'  // темно-фиолетовый
        ];
    }
    
    // Инициализация
    initialize() {
        this.modalElement = document.getElementById('wheelModal');
        
        if (!this.modalElement) {
            console.error('TaskWheel: модальное окно не найдено');
            return;
        }
        
        // Инициализация Bootstrap modal (не закрываемая)
        this.bootstrapModal = new window.bootstrap.Modal(this.modalElement, {
            backdrop: 'static', // Нельзя закрыть кликом по фону
            keyboard: false      // Нельзя закрыть ESC
        });
        
        this.bindEvents();
        console.log('TaskWheel инициализирован');
    }
    
    // Привязка событий
    bindEvents() {
        // Кнопка запуска
        document.getElementById('spinButton').addEventListener('click', () => {
            this.spin();
        });
    }
    
    // Открытие барабана с задачами
    async open() {
        const tasks = await this.tasksService.getAllTasks(`status=open`);

        if (!tasks || tasks.length === 0) {
            alert('Нет доступных задач для выбора');
            return;
        }
        
        this.tasks = tasks;
        this.resetWheel();
        this.renderWheel();
        this.bootstrapModal.show();
    }
    
    // Сброс состояния
    resetWheel() {
        this.currentRotation = 0;
        this.isSpinning = false;
        this.selectedTask = null;
        
        // Сброс UI
        document.getElementById('spinButton').disabled = false;
        document.getElementById('spinButton').hidden = false;
        document.getElementById('spinningIndicator').hidden = true;
        document.getElementById('wheelResult').hidden = true;
        document.getElementById('wheelFooter').hidden = true;
        document.getElementById('wheelSvg').style.transform = 'rotate(0deg)';
        this.clearError();
    }
    
    // Отрисовка барабана
    renderWheel() {
        const segmentsGroup = document.getElementById('wheelSegments');
        segmentsGroup.innerHTML = '';
        
        const numSegments = this.tasks.length;
        const anglePerSegment = 360 / numSegments;
        const radius = 190;
        const centerX = 200;
        const centerY = 200;
        
        this.tasks.forEach((task, index) => {
            const startAngle = index * anglePerSegment;
            const endAngle = (index + 1) * anglePerSegment;
            
            // Создаем сегмент
            const segment = this.createSegment(
                centerX, 
                centerY, 
                radius, 
                startAngle, 
                endAngle, 
                this.colors[index % this.colors.length],
                task.id,
                index
            );
            
            segmentsGroup.appendChild(segment);
        });
    }
    
    // Создание SVG сегмента
    createSegment(cx, cy, radius, startAngle, endAngle, color, taskId, index) {
        const group = document.createElementNS('http://www.w3.org/2000/svg', 'g');
        group.classList.add('wheel-segment');
        group.dataset.taskId = taskId;
        group.dataset.index = index;
        
        // Конвертация углов в радианы
        const startRad = (startAngle - 90) * Math.PI / 180;
        const endRad = (endAngle - 90) * Math.PI / 180;
        
        // Координаты точек
        const x1 = cx + radius * Math.cos(startRad);
        const y1 = cy + radius * Math.sin(startRad);
        const x2 = cx + radius * Math.cos(endRad);
        const y2 = cy + radius * Math.sin(endRad);
        
        // Флаг для больших сегментов
        const largeArcFlag = endAngle - startAngle > 180 ? 1 : 0;
        
        // SVG путь для сегмента
        const pathData = [
            `M ${cx} ${cy}`,
            `L ${x1} ${y1}`,
            `A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2}`,
            'Z'
        ].join(' ');
        
        const path = document.createElementNS('http://www.w3.org/2000/svg', 'path');
        path.setAttribute('d', pathData);
        path.setAttribute('fill', color);
        group.appendChild(path);
        
        // Текст (номер задачи)
        const midAngle = startAngle + (endAngle - startAngle) / 2;
        const textRadius = radius * 0.65;
        const textRad = (midAngle - 90) * Math.PI / 180;
        const textX = cx + textRadius * Math.cos(textRad);
        const textY = cy + textRadius * Math.sin(textRad);
        
        const text = document.createElementNS('http://www.w3.org/2000/svg', 'text');
        text.setAttribute('x', textX);
        text.setAttribute('y', textY);
        text.textContent = `#${taskId}`;
        group.appendChild(text);
        
        return group;
    }
    
    // Вращение барабана
    spin() {
        if (this.isSpinning) return;
        
        this.isSpinning = true;
        
        // Скрываем кнопку, показываем индикатор
        document.getElementById('spinButton').disabled = true;
        document.getElementById('spinningIndicator').hidden = false;
        
        // Случайный угол вращения (минимум 5 полных оборотов + случайный)
        const spins = 5 + Math.random() * 5; // 5-10 оборотов
        const randomAngle = Math.random() * 360;
        const totalRotation = this.currentRotation + (spins * 360) + randomAngle;
        
        // Применяем вращение
        const wheelSvg = document.getElementById('wheelSvg');
        wheelSvg.style.transform = `rotate(${totalRotation}deg)`;
        wheelSvg.classList.add('wheel-spinning');
        
        // Сохраняем текущее вращение
        this.currentRotation = totalRotation;
        
        // Ждем окончания анимации (4 секунды как в CSS)
        setTimeout(() => {
            wheelSvg.classList.remove('wheel-spinning');
            this.determineWinner(totalRotation);
        }, 4000);
    }
    
    // Определение победителя
    determineWinner(finalRotation) {
        const numSegments = this.tasks.length;
        const anglePerSegment = 360 / numSegments;
        
        // Нормализуем угол (0-360)
        const normalizedAngle = finalRotation % 360;
        
        // Вычисляем какой сегмент под указателем (сверху = 0 градусов)
        // Инвертируем потому что вращение по часовой стрелке
        const winningIndex = Math.floor((360 - normalizedAngle) / anglePerSegment) % numSegments;
        
        this.selectedTask = this.tasks[winningIndex];
        
        // Показываем результат
        this.showResult(this.selectedTask);
    }
    
    // Показ результата
    async showResult(task) {

        document.getElementById('wheelFooter').hidden = false;
        const result = await this.handleClaim(task.id);
        document.getElementById('spinningIndicator').hidden = true;
        if(result){
            document.getElementById('wheelResult').hidden = false;
            document.getElementById('resultTaskId').textContent = task.id;
            document.getElementById('resultTaskTitle').textContent = task.name || 'Без названия';
        }
        else{
            setTimeout(() => {
                this.resetWheel();
            }, 5000);
        }

        // Эффект конфетти (опционально)
        //this.celebrate();
    }
    
    // Взять задачу
    async handleClaim(taskId) {
        try {
            const response = await this.tasksService.claimTask(taskId);
            if(response.ok){
                setTimeout(() => {
                    this.bootstrapModal.hide();
                    // Событие для обновления таблицы
                    document.dispatchEvent(new CustomEvent('task:saved'));
                    return true;
                }, 3000);
            }
            else{
                this.showError('Не удалось взять задачу: ' + error.message);
                return false;
            }
        } catch (error) {
            this.showError('Не удалось взять задачу: ' + error.message);
            return false;
        }
    }
    
    // Эффект празднования (простой)
    celebrate() {
        // Можно добавить библиотеку canvas-confetti
        // Или просто анимацию
        const result = document.getElementById('wheelResult');
        result.style.animation = 'bounce 0.5s ease 3';
    }

    showError(message) {
        document.getElementById('wheelErrorText').textContent = message;
        document.getElementById('wheelError').style.display = 'block';
    }

    clearError() {
        document.getElementById('wheelError').style.display = 'none';
    }
}