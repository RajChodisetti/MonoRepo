from __future__ import annotations

from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field


class SceneType(str, Enum):
    electrical_meter_board = "electrical_meter_board"
    gas_infrastructure = "gas_infrastructure"
    mixed = "mixed"
    unknown = "unknown"


class ComponentType(str, Enum):
    electricity_meter = "electricity_meter"
    fuse_holder = "fuse_holder"
    comms_gateway = "comms_gateway"
    antenna = "antenna"
    seal = "seal"
    cable = "cable"
    gas_pipe = "gas_pipe"
    transformer = "transformer"
    conduit = "conduit"
    other = "other"


class Identifiers(BaseModel):
    nmi: Optional[str] = Field(
        default=None,
        description="National Metering Identifier, digits only preferred",
    )
    other_ids: List[str] = Field(default_factory=list)


class MeterInfo(BaseModel):
    brand: Optional[str] = None
    model: Optional[str] = None
    notes: List[str] = Field(default_factory=list)


class FuseInfo(BaseModel):
    rating_amps: Optional[int] = None
    count: int = 1
    label: Optional[str] = None


class ProtectionInfo(BaseModel):
    fuses: List[FuseInfo] = Field(default_factory=list)
    notes: List[str] = Field(default_factory=list)


class Component(BaseModel):
    type: ComponentType = ComponentType.other
    label: str = ""
    confidence: float = Field(default=0.0, ge=0.0, le=1.0)


class UnlabeledDetection(BaseModel):
    object: str
    evidence: str = ""
    confidence: float = Field(default=0.0, ge=0.0, le=1.0)


class ElectricalAnalysis(BaseModel):
    """Structured extraction from an electrical / utility site photo."""

    scene_type: SceneType = SceneType.unknown
    summary: str = ""
    identifiers: Identifiers = Field(default_factory=Identifiers)
    meter: MeterInfo = Field(default_factory=MeterInfo)
    protection: ProtectionInfo = Field(default_factory=ProtectionInfo)
    components: List[Component] = Field(default_factory=list)
    handwritten_labels: List[str] = Field(default_factory=list)
    warnings: List[str] = Field(default_factory=list)
    unlabeled_detections: List[UnlabeledDetection] = Field(default_factory=list)
    raw_ocr_text: str = ""
