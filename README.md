# voca-video-maker
영어 단어 영상 서비스

## 기능

### 1. 기존 기능
- 영어 단어와 한국어 번역으로 이미지 생성
- 음성 파일 생성 (한국어/영어)
- 이미지와 음성을 합쳐서 영상 생성
- 여러 영상을 하나로 합치기

- **해상도**: 1080x1920 (9:16 비율)
- **비디오 코덱**: H.264 Baseline Profile
- **오디오 코덱**: AAC
- **오디오 샘플레이트**: 44.1kHz
- **비디오 품질**: CRF 25 (모바일 최적화)

### FFmpeg 최적화

- `-preset fast`: 빠른 인코딩
- `-profile:v baseline`: 모바일 호환성 최대화
- `-level 3.0`: 널리 지원되는 레벨
- `-crf 25`: 모바일 최적 품질
- `-vf scale=1080:1920`: 세로 비디오로 리사이즈
- `-ar 44100`: 44.1kHz 오디오 샘플레이트
- `-movflags +faststart`: 스트리밍 최적화
- `-avoid_negative_ts make_zero`: 정확한 싱크
- `-fflags +genpts`: Presentation Time Stamp 재생성

## 요구사항

- Go 1.16 이상
- FFmpeg 설치
- macOS (say 명령어 사용)
